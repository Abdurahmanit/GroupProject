package mailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMailer для подмены реальной отправки писем
type MockMailer struct {
	WasCalled bool
}

func (m *MockMailer) SendListingCreatedEmail(toEmail, listingTitle string) error {
	m.WasCalled = true
	return nil // эмулируем успешную отправку
}

func TestSendListingCreatedEmail_Mock(t *testing.T) {
	mock := &MockMailer{}
	err := mock.SendListingCreatedEmail("test@example.com", "Test Listing")

	assert.NoError(t, err)
	assert.True(t, mock.WasCalled)
}
