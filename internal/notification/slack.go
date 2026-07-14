package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type SlackNotification struct {
	WebhookURL string
	Message    string
	AlarmID    string
}

func (s *SlackNotification) Build(alarmConfig *alarm.Alarm, sig signal.Signal) error {
	s.AlarmID = alarmConfig.ID
	s.Message = fmt.Sprintf("*Alert:* %s (*%s*)\n*Signal:* %v", alarmConfig.Name, sig.Status, sig)
	return nil
}

func (s *SlackNotification) Send() error {
	payload := map[string]string{
		"text": s.Message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", s.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Printf("slack notification sending alarm_id=%s", s.AlarmID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-success response from Slack: %s", resp.Status)
	}
	log.Printf("slack notification response alarm_id=%s status=%s", s.AlarmID, resp.Status)

	return nil
}

func NewSlackNotification(webhookURL string) *SlackNotification {
	return &SlackNotification{WebhookURL: webhookURL}
}
