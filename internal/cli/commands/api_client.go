package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/g0ulartleo/mirante-alerts/internal/alarm"
	"github.com/g0ulartleo/mirante-alerts/internal/config"
	"github.com/g0ulartleo/mirante-alerts/internal/signal"
)

type Client struct {
	config *config.CLIConfig
}

func NewAPIClient(config *config.CLIConfig) *Client {
	return &Client{config: config}
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

	apiHost := c.config.APIHost
	if apiHost != "" && !hasScheme(apiHost) {
		apiHost = "http://" + apiHost
	}

	url := apiHost + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	switch c.config.AuthType {
	case "oauth":
		if c.config.AuthToken == "" {
			return nil, fmt.Errorf("no OAuth token configured. Run './cli auth <api_host>' to authenticate")
		}
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	case "api_key":
		if c.config.APIKey == "" {
			return nil, fmt.Errorf("no API key configured. Run './cli config <api_host> <api_key>' to configure")
		}
		req.Header.Set("X-API-Key", c.config.APIKey)
	default:
		if c.config.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
		} else if c.config.APIKey != "" {
			req.Header.Set("X-API-Key", c.config.APIKey)
		} else {
			return nil, fmt.Errorf("no authentication configured. Run 'mirante auth-key <api_host> <api_key>' or 'mirante auth <api_host>'")
		}
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

func (c *Client) DeleteAlarm(id string) error {
	endpoint := path.Join("/api/alarms", id)
	_, err := c.doRequest(http.MethodDelete, endpoint, nil)
	return err
}

func (c *Client) SetAlarm(a *alarm.Alarm) error {
	_, err := c.doRequest(http.MethodPost, "/api/alarms", a)
	return err
}

func (c *Client) GetAlarmSignals(id string) ([]signal.Signal, error) {
	endpoint := path.Join("/api/alarms", id, "signals")
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

func (c *Client) CheckAlarm(id string) error {
	endpoint := path.Join("/api/alarms", id, "check")
	_, err := c.doRequest(http.MethodPost, endpoint, nil)
	return err
}

func hasScheme(urlStr string) bool {
	return len(urlStr) > 7 && (urlStr[:7] == "http://" || urlStr[:8] == "https://")
}
