package email

import (
	"testing"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSMTPSender_SendEmail_IncompleteConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := []struct {
		name        string
		cfg         *config.SMTPConfig
		expectedErr string
	}{
		{
			name: "Missing Username",
			cfg: &config.SMTPConfig{
				Host:        "smtp.example.com",
				Port:        587,
				Password:    "fakepassword",
				SenderEmail: "sender@example.com",
			},
			expectedErr: "SMTP configuration is incomplete",
		},
		{
			name: "Missing Host",
			cfg: &config.SMTPConfig{
				Port:        587,
				Username:    "user",
				Password:    "fakepassword",
				SenderEmail: "sender@example.com",
			},
			expectedErr: "SMTP configuration is incomplete",
		},
		{
			name: "Missing Password",
			cfg: &config.SMTPConfig{
				Host:        "smtp.example.com",
				Port:        587,
				Username:    "user",
				SenderEmail: "sender@example.com",
			},
			expectedErr: "SMTP configuration is incomplete",
		},
		{
			name: "Missing SenderEmail",
			cfg: &config.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "fakepassword",
			},
			expectedErr: "SMTP configuration is incomplete",
		},
		{
			name:        "All Missing",
			cfg:         &config.SMTPConfig{},
			expectedErr: "SMTP configuration is incomplete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ВАЖНО: Убедитесь, что имя функции здесь (NewSMTPSender или newSMTPSender)
			// точно совпадает с именем в вашем файле smtp_sender.go
			senderInstance := NewSMTPSender(tc.cfg, logger) // ИЛИ newSMTPSender, если так у вас

			err := senderInstance.SendEmail([]string{"recipient@example.com"}, "Test Subject", "Test Body")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}
