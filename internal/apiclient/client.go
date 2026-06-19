package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	"golang.org/x/net/websocket"
)

type Client struct {
	config *config.CLIConfig
}

func New(cfg *config.CLIConfig) *Client {
	return &Client{config: cfg}
}

func (c *Client) doRequest(method, endpoint string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	u := c.apiHost() + endpoint
	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if err := c.setAuthHeader(req.Header); err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		if len(errorBody) > 0 {
			return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(errorBody))
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) apiHost() string {
	apiHost := c.config.APIHost
	if apiHost != "" && !hasScheme(apiHost) {
		apiHost = "http://" + apiHost
	}
	return strings.TrimRight(apiHost, "/")
}

func (c *Client) setAuthHeader(h http.Header) error {
	switch c.config.AuthType {
	case "oauth":
		if c.config.AuthToken == "" {
			return fmt.Errorf("no OAuth token configured. Run 'mirante auth <api_host>' to authenticate")
		}
		h.Set("Authorization", "Bearer "+c.config.AuthToken)
	case "api_key":
		if c.config.APIKey == "" {
			return fmt.Errorf("no API key configured. Run 'mirante auth-key <api_host> <api_key>' to configure")
		}
		h.Set("X-API-Key", c.config.APIKey)
	default:
		if c.config.AuthToken != "" {
			h.Set("Authorization", "Bearer "+c.config.AuthToken)
		} else if c.config.APIKey != "" {
			h.Set("X-API-Key", c.config.APIKey)
		} else {
			return fmt.Errorf("no authentication configured. Run 'mirante auth-key <api_host> <api_key>' or 'mirante auth <api_host>'")
		}
	}
	return nil
}

func (c *Client) ListAlarms() ([]alarm.Alarm, error) {
	data, err := c.doRequest(http.MethodGet, "/api/alarms", nil)
	if err != nil {
		return nil, err
	}

	var alarms []alarm.Alarm
	if err := json.Unmarshal(data, &alarms); err != nil {
		return nil, err
	}

	return alarms, nil
}

func (c *Client) GetAlarm(id string) (*alarm.Alarm, error) {
	endpoint := path.Join("/api/alarms", id)
	data, err := c.doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var a alarm.Alarm
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}

	return &a, nil
}

func (c *Client) GetAlarmSignals(id string, limit int) ([]signal.Signal, error) {
	endpoint := path.Join("/api/alarms", id, "signals")
	if limit > 0 {
		endpoint += "?limit=" + strconv.Itoa(limit)
	}
	data, err := c.doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var signals []signal.Signal
	if err := json.Unmarshal(data, &signals); err != nil {
		return nil, err
	}

	return signals, nil
}

func (c *Client) GetAlarmSignalsSince(id string, since time.Time) ([]signal.Signal, error) {
	endpoint := path.Join("/api/alarms", id, "signals")
	query := url.Values{}
	query.Set("since", since.Format(time.RFC3339))
	endpoint += "?" + query.Encode()

	data, err := c.doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var signals []signal.Signal
	if err := json.Unmarshal(data, &signals); err != nil {
		return nil, err
	}

	return signals, nil
}

func (c *Client) GetAllAlarmSignals() ([]alarm.AlarmSignals, error) {
	data, err := c.doRequest(http.MethodGet, "/api/alarms/signals", nil)
	if err != nil {
		return nil, err
	}

	var alarmSignals []alarm.AlarmSignals
	if err := json.Unmarshal(data, &alarmSignals); err != nil {
		return nil, err
	}

	return alarmSignals, nil
}

func (c *Client) RunAlarm(id string) error {
	endpoint := path.Join("/api/alarms", id, "check")
	_, err := c.doRequest(http.MethodPost, endpoint, nil)
	return err
}

func (c *Client) SyncAlarms() error {
	_, err := c.doRequest(http.MethodPost, "/api/alarms/sync", nil)
	return err
}

func (c *Client) SubscribeAlarmSignals(ctx context.Context) (<-chan []alarm.AlarmSignals, <-chan error, error) {
	wsURL, err := c.websocketURL("/api/alarms/ws")
	if err != nil {
		return nil, nil, err
	}

	cfg, err := websocket.NewConfig(wsURL, c.apiHost())
	if err != nil {
		return nil, nil, err
	}
	if err := c.setAuthHeader(cfg.Header); err != nil {
		return nil, nil, err
	}

	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %w", wsURL, err)
	}

	updates := make(chan []alarm.AlarmSignals)
	errs := make(chan error, 1)

	go func() {
		defer close(updates)
		defer conn.Close()

		go func() {
			<-ctx.Done()
			conn.Close()
		}()

		for {
			var payload []byte
			if err := websocket.Message.Receive(conn, &payload); err != nil {
				if ctx.Err() == nil {
					select {
					case errs <- err:
					default:
					}
				}
				return
			}

			var alarmSignals []alarm.AlarmSignals
			if err := json.Unmarshal(payload, &alarmSignals); err != nil {
				select {
				case errs <- err:
				default:
				}
				continue
			}

			select {
			case updates <- alarmSignals:
			case <-ctx.Done():
				return
			}
		}
	}()

	return updates, errs, nil
}

func (c *Client) websocketURL(endpoint string) (string, error) {
	u, err := url.Parse(c.apiHost() + endpoint)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}
	return u.String(), nil
}

func hasScheme(urlStr string) bool {
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}
