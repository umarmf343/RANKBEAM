package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Mailer interface {
	SendLicenseEmail(ctx context.Context, to string, license *License) error
}

type SMTPMailer struct {
	addr   string
	host   string
	from   string
	auth   smtp.Auth
	tlsCfg *tls.Config
}

func NewSMTPMailer(host string, port int, username, password, from string) (*SMTPMailer, error) {
	host = strings.TrimSpace(host)
	from = strings.TrimSpace(from)
	if host == "" {
		return nil, errors.New("mailer: host is required")
	}
	if from == "" {
		return nil, errors.New("mailer: from address is required")
	}
	if port <= 0 {
		port = 465
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	var auth smtp.Auth
	if strings.TrimSpace(username) != "" && strings.TrimSpace(password) != "" {
		auth = smtp.PlainAuth("", strings.TrimSpace(username), strings.TrimSpace(password), host)
	}

	tlsCfg := &tls.Config{ServerName: host}

	return &SMTPMailer{addr: addr, host: host, from: from, auth: auth, tlsCfg: tlsCfg}, nil
}

func (m *SMTPMailer) SendLicenseEmail(ctx context.Context, to string, license *License) error {
	if m == nil {
		return errors.New("mailer: smtp mailer is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	to = strings.TrimSpace(strings.ToLower(to))
	if to == "" {
		return errors.New("mailer: recipient email is required")
	}
	if license == nil {
		return errors.New("mailer: license payload is required")
	}

	subject := "Your RankBeam license key"
	expiry := "never"
	if !license.ExpiresAt.IsZero() {
		expiry = license.ExpiresAt.UTC().Format(time.RFC1123)
	}

	body := fmt.Sprintf("Hello,\n\nThank you for keeping your RankBeam subscription active. Here is your current license key (valid until %s UTC):\n\n%s\n\nYou can paste this key inside the RankBeam installer or the desktop app's activation screen.\n\nIf you did not expect this email, please contact support immediately.\n\n— RankBeam Support\n", expiry, license.Key)

	headers := map[string]string{
		"From":         m.from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=utf-8",
	}

	var msgBuilder strings.Builder
	for k, v := range headers {
		msgBuilder.WriteString(k)
		msgBuilder.WriteString(": ")
		msgBuilder.WriteString(v)
		msgBuilder.WriteString("\r\n")
	}
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)
	message := []byte(msgBuilder.String())

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.send(message, to)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (m *SMTPMailer) send(message []byte, to string) error {
	conn, err := tls.Dial("tcp", m.addr, m.tlsCfg)
	if err != nil {
		return fmt.Errorf("mailer: dial smtp: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("mailer: create smtp client: %w", err)
	}
	defer client.Close()

	if m.auth != nil {
		if err := client.Auth(m.auth); err != nil {
			return fmt.Errorf("mailer: authenticate: %w", err)
		}
	}

	if err := client.Mail(m.from); err != nil {
		return fmt.Errorf("mailer: set from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("mailer: set recipient: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: get data writer: %w", err)
	}
	if _, err := wc.Write(message); err != nil {
		wc.Close()
		return fmt.Errorf("mailer: write message: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mailer: close writer: %w", err)
	}
	return client.Quit()
}
