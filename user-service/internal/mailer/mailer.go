package mailer

// Mailer defines the interface for sending emails.
type Mailer interface {
	SendEmailVerification(toEmail, toName, verificationCode string) error
}
