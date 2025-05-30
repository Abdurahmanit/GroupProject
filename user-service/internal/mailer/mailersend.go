package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const mailerSendAPIURL = "https://api.mailersend.com/v1/email"

// MailerSendService implements the Mailer interface using MailerSend.
type MailerSendService struct {
	apiKey    string
	fromEmail string
	fromName  string
	client    *http.Client
	logger    *zap.Logger
}

// NewMailerSendService creates a new MailerSendService.
func NewMailerSendService(apiKey, fromEmail, fromName string, logger *zap.Logger) *MailerSendService {
	return &MailerSendService{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.Named("MailerSendService"),
	}
}

type mailerSendRequest struct {
	From            fromEmail              `json:"from"`
	To              []toEmail              `json:"to"`
	Subject         string                 `json:"subject"`
	Text            string                 `json:"text"`
	HTML            string                 `json:"html"`
	Personalization []personalizationEntry `json:"personalization,omitempty"`
}

type fromEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type toEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type personalizationEntry struct {
	Email string            `json:"email"`
	Data  map[string]string `json:"data"`
}

// SendEmailVerification sends a verification email to the user.
func (s *MailerSendService) SendEmailVerification(toEmailAddr, toName, verificationCode string) error {
	s.logger.Info("Attempting to send verification email", zap.String("toEmail", toEmailAddr))

	subject := "Verify Your Email Address"
	htmlBody := fmt.Sprintf(`<p>Hello %s,</p>
                             <p>Your verification code is: <b>%s</b></p>
                             <p>This code will expire in 15 minutes.</p>
                             <p>If you did not request this, please ignore this email.</p>`, toName, verificationCode)
	textBody := fmt.Sprintf(`Hello %s,
                           Your verification code is: %s
                           This code will expire in 15 minutes.
                           If you did not request this, please ignore this email.`, toName, verificationCode)

	requestPayload := mailerSendRequest{
		From: fromEmail{
			Email: s.fromEmail,
			Name:  s.fromName,
		},
		To: []toEmail{
			{Email: toEmailAddr, Name: toName},
		},
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
		Personalization: []personalizationEntry{
			{
				Email: toEmailAddr,
				Data: map[string]string{
					"name": toName,
					"code": verificationCode,
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		s.logger.Error("Failed to marshal MailerSend request payload", zap.Error(err))
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", mailerSendAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Error("Failed to create MailerSend HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request to MailerSend", zap.Error(err))
		return fmt.Errorf("failed to send request to MailerSend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		s.logger.Error("MailerSend API request failed", zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("MailerSend API request failed with status code %d", resp.StatusCode)
	}

	s.logger.Info("Verification email sent successfully via MailerSend", zap.String("toEmail", toEmailAddr), zap.String("messageID", resp.Header.Get("X-Message-Id")))
	return nil
}
