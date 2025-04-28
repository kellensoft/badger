package auth

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailSender() *EmailSender {
	return &EmailSender{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM_EMAIL"),
	}
}

func (e *EmailSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", e.host, e.port)

	auth := smtp.PlainAuth("", e.username, e.password, e.host)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", e.from, to, subject, body))

	err := smtp.SendMail(addr, auth, e.from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
