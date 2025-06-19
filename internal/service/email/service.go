// Package email provides email sending functionality for the authentication service.
package email

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/gofiber/fiber/v2/log"
)

// Service provides email sending functionalities.
type Service struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(host, port, username, password string) *Service {
	return &Service{
		SMTPHost: host,
		SMTPPort: port,
		Username: username,
		Password: password,
	}
}

// SendEmailHTML sends an HTML email.
func (s *Service) SendEmailHTML(to, subject, htmlBody string) error {
	from := s.Username
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=UTF-8\n\n" +
		htmlBody

	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)
	err := smtp.SendMail(s.SMTPHost+":"+s.SMTPPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Error("Error sending HTML email: %v", err)
		return err
	}
	return nil
}

// SendPasswordReset sends a password reset email.
func (s *Service) SendPasswordReset(email, link string, expireMinutes int) error {
	tmpl, err := template.New("reset").Parse(passwordResetEmailTemplate)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, PasswordResetEmailData{
		ResetLink:     link,
		ExpireMinutes: expireMinutes,
	})
	if err != nil {
		return err
	}
	subject := "[YourApp] 비밀번호 재설정 안내"
	return s.SendEmailHTML(email, subject, body.String())
}
