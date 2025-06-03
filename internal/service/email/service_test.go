package email_test

import (
	"auth/internal/config"
	"auth/internal/service/email"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_메일HTML전송실패(t *testing.T) {
	// Initialize the email service with test credentials
	emailService := email.NewEmailService("smtp.example.com", "587", "test@example.com", "testpassword")
	err := emailService.SendEmailHtml("invalid-email", "Test HTML Subject", "<h1>Test HTML Body</h1>")
	assert.NotNil(t, err, "Expected an error but got nil")
}

func Test_메일HTML전송성공(t *testing.T) {
	// Initialize the email service with test credentials
	config := config.LoadConfig("E:/workspace/auth/.env")
	emailService := email.NewEmailService(config.SmtpServer, config.SmtpPort, config.SmtpId, config.SmtpPassword)
	resetLink := fmt.Sprintf("https://yourdomain.com/reset-password?token=%s", "21312312312")
	err := emailService.SendPasswordReset("dgkwon90@naver.com", resetLink, 30)
	assert.Nil(t, err, "Expected no error but got one")
}
