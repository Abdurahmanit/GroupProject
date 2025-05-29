package mailer

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"os"
)

type Mailer interface {
	SendListingCreatedEmail(toEmail, listingTitle string) error
}

type SMTPMailer struct{}

func SendListingCreatedEmail(toEmail, listingTitle string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	if from == "" || password == "" {
		return fmt.Errorf("SMTP credentials not set")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "New Listing Created")
	m.SetBody("text/plain", "Your listing '"+listingTitle+"' has been created successfully.")

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)
	return d.DialAndSend(m)
}
