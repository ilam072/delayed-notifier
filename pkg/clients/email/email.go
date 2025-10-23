package email

import (
	"fmt"
	"net"
	"net/smtp"
)

type Config struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
	From     string
}

type Client struct {
	cfg Config
}

// New создаёт новый email-клиент.
func New(smtpHost, smtpPort, username, password, from string) *Client {
	cfg := Config{
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		Username: username,
		Password: password,
		From:     from,
	}
	return &Client{cfg: cfg}
}

// Send отправляет email с фиксированной темой "Notification" и переданным текстом.
func (c Client) Send(message, recipient string) error {
	addr := net.JoinHostPort(c.cfg.SMTPHost, c.cfg.SMTPPort)

	subject := "Notification"

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		c.cfg.From, recipient, subject, message,
	))

	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.SMTPHost)
	to := []string{recipient}

	return smtp.SendMail(addr, auth, c.cfg.From, to, msg)
}
