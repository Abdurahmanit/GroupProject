package mailer

import (
	"gopkg.in/gomail.v2"
	"os"
)

func SendListingCreatedEmail(toEmail, listingTitle string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "New Listing Created")
	m.SetBody("text/plain", "Your listing '"+listingTitle+"' has been created successfully.")

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)
	return d.DialAndSend(m)
}
