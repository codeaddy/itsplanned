package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

var config EmailConfig

func Init() error {
	config = EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
	}

	if config.SMTPHost == "" || config.SMTPPort == "" ||
		config.SMTPUsername == "" || config.SMTPPassword == "" ||
		config.FromEmail == "" {
		return fmt.Errorf("missing required email configuration")
	}

	return nil
}

func SendPasswordResetEmail(toEmail, resetToken string) error {
	tlsConfig := &tls.Config{
		ServerName: config.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", config.SMTPHost+":"+config.SMTPPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %v", err)
	}
	defer writer.Close()

	subject := "Password Reset Request"
	resetLink := fmt.Sprintf("http://localhost:8080/password/reset-redirect?token=%s", resetToken)
	body := fmt.Sprintf(`
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body>
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password. Click the link below to proceed:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>If you did not request this password reset, please ignore this email.</p>
			<p>This link will expire in 15 minutes.</p>
		</body>
		</html>
	`, resetLink)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", config.FromEmail, toEmail, subject, body)

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email message: %v", err)
	}

	return nil
}
