package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port      int    `mapstructure:"PORT"`
	MongoURI  string `mapstructure:"MONGO_URI"`
	RedisAddr string `mapstructure:"REDIS_ADDR"`
	JWTSecret string `mapstructure:"JWT_SECRET"`

	MailerType string `mapstructure:"MAILER_TYPE"` // "mailersend" or "smtp"

	// MailerSend specific
	MailerSendAPIKey    string `mapstructure:"MAILERSEND_API_KEY"`
	MailerSendFromEmail string `mapstructure:"MAILERSEND_FROM_EMAIL"` // For MailerSend
	MailerSendFromName  string `mapstructure:"MAILERSEND_FROM_NAME"`  // For MailerSend

	// SMTP specific
	SMTPHost       string `mapstructure:"SMTP_HOST"`
	SMTPPort       int    `mapstructure:"SMTP_PORT"`
	SMTPUsername   string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword   string `mapstructure:"SMTP_PASSWORD"`
	SMTPFromEmail  string `mapstructure:"SMTP_FROM_EMAIL"`  // For SMTP, often same as username
	SMTPSenderName string `mapstructure:"SMTP_SENDER_NAME"` // For SMTP
}

func LoadConfig() (*Config, error) {
	// Bind common environment variables
	viper.BindEnv("port", "PORT")
	viper.BindEnv("mongo_uri", "MONGO_URI")
	viper.BindEnv("redis_addr", "REDIS_ADDR")
	viper.BindEnv("jwt_secret", "JWT_SECRET")
	viper.BindEnv("mailer_type", "MAILER_TYPE")

	// Bind MailerSend specific
	viper.BindEnv("mailersend_api_key", "MAILERSEND_API_KEY")
	viper.BindEnv("mailersend_from_email", "MAILERSEND_FROM_EMAIL")
	viper.BindEnv("mailersend_from_name", "MAILERSEND_FROM_NAME")

	// Bind SMTP specific
	viper.BindEnv("smtp_host", "SMTP_HOST")
	viper.BindEnv("smtp_port", "SMTP_PORT")
	viper.BindEnv("smtp_username", "SMTP_USERNAME")
	viper.BindEnv("smtp_password", "SMTP_PASSWORD")
	viper.BindEnv("smtp_from_email", "SMTP_FROM_EMAIL")
	viper.BindEnv("smtp_sender_name", "SMTP_SENDER_NAME")

	viper.AutomaticEnv() // Allow Viper to pick up other env vars

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Set a default mailer type if not specified
	if cfg.MailerType == "" {
		cfg.MailerType = "mailersend" // Or "smtp" depending on your primary choice
	}

	return &cfg, nil
}
