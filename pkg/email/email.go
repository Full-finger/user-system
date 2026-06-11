// Package email 提供 SMTP 邮件发送功能，支持 TLS 和明文两种模式。
package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// Sender SMTP 邮件发送器，实现 service.Mailer 接口。
type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	tls      bool
	auth     bool
}

func NewSender(host string, port int, username, password, from string, tls, auth bool) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		tls:      tls,
		auth:     auth,
	}
}

// sanitizeHeader 移除字符串中的 \r 和 \n，防止邮件头注入。
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// Send 发送纯文本邮件。
func (s *Sender) Send(to, subject, body string) error {
	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))

	msg := "From: " + sanitizeHeader(s.from) + "\r\n" +
		"To: " + sanitizeHeader(to) + "\r\n" +
		"Subject: " + sanitizeHeader(subject) + "\r\n" +
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

		if s.auth {
			smtpAuth := smtp.PlainAuth("", s.username, s.password, s.host)
			if err = client.Auth(smtpAuth); err != nil {
				return fmt.Errorf("smtp auth failed: %w", err)
			}
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

	if s.auth {
		smtpAuth := smtp.PlainAuth("", s.username, s.password, s.host)
		return smtp.SendMail(addr, smtpAuth, s.from, []string{to}, []byte(msg))
	}
	return smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg))
}
