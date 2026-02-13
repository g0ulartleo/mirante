package notification

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type EmailNotification struct {
	To      []string
	Subject string
	Body    string
}

func (e *EmailNotification) Build(alarmConfig *alarm.Alarm, sig signal.Signal) error {
	e.To = alarmConfig.Notifications.Email.To
	e.Subject = fmt.Sprintf("%s is %s", alarmConfig.Name, sig.Status)
	e.Body = sig.Message
	return nil
}

func (e *EmailNotification) Send() error {
	env := config.Env()
	from := env.SMTPUser
	password := env.SMTPPassword
	smtpHost := env.SMTPHost
	smtpPort := env.SMTPPort
	to := e.To

	message := []byte("From: " + from + "\r\n" +
		"To: " + strings.Join(to, ",") + "\r\n" +
		"Subject: " + e.Subject + "\r\n" +
		"\r\n" +
		e.Body)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":1"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}

func NewEmailNotification() *EmailNotification {
	return &EmailNotification{}
}
