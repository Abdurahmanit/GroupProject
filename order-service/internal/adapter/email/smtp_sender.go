package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	"gopkg.in/gomail.v2"
)

type EmailSender interface {
	Send(ctx context.Context, to []string, subject, bodyHTML, bodyText string) error
}

type smtpSender struct {
	cfg config.SMTPConfig
	log logger.Logger
	d   *gomail.Dialer
}

func NewSMTPSender(cfg config.SMTPConfig, log logger.Logger) (EmailSender, error) {
	if cfg.Host == "" || cfg.Port == 0 || cfg.SenderEmail == "" {
		return nil, fmt.Errorf("SMTP host, port, and sender email must be configured")
	}

	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	encryptionLower := strings.ToLower(cfg.Encryption)
	serverName := cfg.ServerName
	if serverName == "" {
		serverName = cfg.Host
	}

	if encryptionLower == "ssl" {
		dialer.SSL = true
		dialer.TLSConfig = &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12}
	} else if encryptionLower == "tls" || encryptionLower == "starttls" {
		if dialer.TLSConfig == nil {
			dialer.TLSConfig = &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12}
		} else {
			dialer.TLSConfig.ServerName = serverName
			dialer.TLSConfig.MinVersion = tls.VersionTLS12
		}
	}

	return &smtpSender{
		cfg: cfg,
		log: log,
		d:   dialer,
	}, nil
}

func (s *smtpSender) Send(ctx context.Context, to []string, subject, bodyHTML, bodyText string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients provided for email")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SenderEmail)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	if bodyHTML != "" {
		m.SetBody("text/html", bodyHTML)
		if bodyText != "" {
			m.AddAlternative("text/plain", bodyText)
		}
	} else if bodyText != "" {
		m.SetBody("text/plain", bodyText)
	} else {
		return fmt.Errorf("email body (HTML or Text) must be provided")
	}

	var err error
	sendAttempt := func() error {
		return s.d.DialAndSend(m)
	}

	done := make(chan error, 1)
	go func() {
		done <- sendAttempt()
	}()

	select {
	case <-ctx.Done():
		s.log.Warnf("Email sending to %v (subject: %s) cancelled or timed out by context: %v", to, subject, ctx.Err())
		return fmt.Errorf("email sending cancelled or timed out: %w", ctx.Err())
	case err = <-done:
		if err != nil {
			s.log.Errorf("Failed to send email to %v, subject '%s': %v", to, subject, err)
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	s.log.Infof("Email sent successfully to %v, subject: %s", to, subject)
	return nil
}
