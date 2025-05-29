package mailer

import (
	"os"
	"testing"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env") // Путь до .env из internal/mailer
	if err != nil {
		panic("Не удалось загрузить .env: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestSendListingCreatedEmail_Integration(t *testing.T) {
	m := &SMTPMailer{}

	to := os.Getenv("TEST_RECEIVER_EMAIL")
	if to == "" {
		t.Skip("TEST_RECEIVER_EMAIL не задан — пропуск интеграционного теста")
	}

	err := m.SendListingCreatedEmail(to, "Integration Test Listing")
	if err != nil {
		t.Errorf("Не удалось отправить email: %v", err)
	}
}
