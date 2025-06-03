package email

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/gofiber/fiber/v2/log"
)

type EmailService struct {
	SmtpHost string
	SmtpPort string
	Username string
	Password string
}

func NewEmailService(host, port, username, password string) *EmailService {
	return &EmailService{
		SmtpHost: host,
		SmtpPort: port,
		Username: username,
		Password: password,
	}
}

func (s *EmailService) SendEmailHtml(to, subject, htmlBody string) error {
	from := s.Username
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=UTF-8\n\n" +
		htmlBody

	auth := smtp.PlainAuth("", s.Username, s.Password, s.SmtpHost)
	err := smtp.SendMail(s.SmtpHost+":"+s.SmtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Error("Error sending HTML email: %v", err)
		return err
	}
	return nil
}

func (s *EmailService) SendPasswordReset(email, link string, expireMinutes int) error {
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
	return s.SendEmailHtml(email, subject, body.String())
}
