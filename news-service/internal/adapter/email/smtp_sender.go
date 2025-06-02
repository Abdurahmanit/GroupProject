package email

import (
	"fmt"
	"net/smtp"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"go.uber.org/zap"
)

type Sender interface {
	SendEmail(to []string, subject, body string) error
}

type smtpSender struct {
	cfg    *config.SMTPConfig
	logger *zap.Logger
}

func NewSMTPSender(cfg *config.SMTPConfig, logger *zap.Logger) Sender {
	return &smtpSender{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *smtpSender) SendEmail(to []string, subject, body string) error {
	if s.cfg.Host == "" || s.cfg.Username == "" || s.cfg.Password == "" || s.cfg.SenderEmail == "" {
		s.logger.Error("SMTP configuration is incomplete. Email not sent.",
			zap.String("host", s.cfg.Host),
			zap.String("username", s.cfg.Username),
			zap.Bool("password_set", s.cfg.Password != ""),
			zap.String("sender", s.cfg.SenderEmail))
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	toList := ""
	for i, recipient := range to {
		toList += recipient
		if i < len(to)-1 {
			toList += ","
		}
	}

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.cfg.SenderEmail, toList, subject, body))

	err := smtp.SendMail(addr, auth, s.cfg.SenderEmail, to, msg)
	if err != nil {
		s.logger.Error("Failed to send email", zap.Error(err), zap.Strings("to", to), zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully", zap.Strings("to", to), zap.String("subject", subject))
	return nil
}
