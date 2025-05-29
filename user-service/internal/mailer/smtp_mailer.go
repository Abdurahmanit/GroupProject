package mailer

import (
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// SMTPMailerService implements the Mailer interface using net/smtp.
type SMTPMailerService struct {
	host       string
	port       int
	username   string
	password   string
	from       string // The "From" address for the email header
	senderName string // The display name for the sender
	logger     *zap.Logger
}

// NewSMTPMailerService creates a new SMTPMailerService.
func NewSMTPMailerService(host string, port int, username, password, fromEmail, senderName string, logger *zap.Logger) *SMTPMailerService {
	return &SMTPMailerService{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		from:       fromEmail,
		senderName: senderName,
		logger:     logger.Named("SMTPMailerService"),
	}
}

// SendEmailVerification sends a verification email using SMTP.
func (s *SMTPMailerService) SendEmailVerification(toEmailAddr, toName, verificationCode string) error {
	s.logger.Info("Attempting to send verification email via SMTP",
		zap.String("toEmail", toEmailAddr),
		zap.String("smtpHost", s.host),
		zap.Int("smtpPort", s.port))

	subject := "Verify Your Email Address"

	// Construct email body (HTML and Plain Text)
	// For simplicity, reusing the same content logic
	htmlBodyContent := fmt.Sprintf(`<p>Hello %s,</p>
                             <p>Your verification code is: <b>%s</b></p>
                             <p>This code will expire in 15 minutes.</p>
                             <p>If you did not request this, please ignore this email.</p>`, toName, verificationCode)

	plainTextBodyContent := fmt.Sprintf(`Hello %s,
                           Your verification code is: %s
                           This code will expire in 15 minutes.
                           If you did not request this, please ignore this email.`, toName, verificationCode)

	// Setup SMTP authentication
	// The address for smtp.PlainAuth should be "host:port", but for smtp.SendMail it's just "host:port"
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Email headers
	headers := make(map[string]string)
	if s.senderName != "" {
		headers["From"] = fmt.Sprintf("%s <%s>", s.senderName, s.from)
	} else {
		headers["From"] = s.from
	}
	headers["To"] = toEmailAddr
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"

	// Constructing a multipart message
	boundary := "my-boundary-12345" // Can be any unique string
	headers["Content-Type"] = fmt.Sprintf("multipart/alternative; boundary=%s", boundary)

	var msgBuilder strings.Builder

	// Write headers
	for k, v := range headers {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msgBuilder.WriteString("\r\n") // Empty line separates headers from body

	// Plain text part
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msgBuilder.WriteString(plainTextBodyContent)
	msgBuilder.WriteString("\r\n\r\n")

	// HTML part
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msgBuilder.WriteString(htmlBodyContent)
	msgBuilder.WriteString("\r\n\r\n")

	// End boundary
	msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	msg := msgBuilder.String()

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Send the email
	err := smtp.SendMail(addr, auth, s.from, []string{toEmailAddr}, []byte(msg))
	if err != nil {
		s.logger.Error("Failed to send email via SMTP",
			zap.Error(err),
			zap.String("toEmail", toEmailAddr),
			zap.String("smtpHost", s.host))
		return fmt.Errorf("smtp.SendMail failed: %w", err)
	}

	s.logger.Info("Verification email sent successfully via SMTP", zap.String("toEmail", toEmailAddr))
	return nil
}
