package pkg

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type EmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	tls      bool
}

func NewEmailSender(host string, port int, username, password, from string, tls bool) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		tls:      tls,
	}
}

func (s *EmailSender) Send(to, subject, body string) error {
	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	if s.tls {
		tlsConfig := &tls.Config{ServerName: s.host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial failed: %w", err)
		}
		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("create smtp client failed: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
		if err = client.Mail(s.from); err != nil {
			return fmt.Errorf("smtp mail from failed: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt to failed: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data failed: %w", err)
		}
		if _, err = w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("write message failed: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("close writer failed: %w", err)
		}
		return client.Quit()
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}
