// Package email_test contains unit tests for the email service, including HTML email and password reset scenarios.
package email_test

import (
	"auth/internal/config"
	"auth/internal/service/email"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSendEmailHTMLFail tests that sending an email with an invalid address returns an error.
func TestSendEmailHTMLFail(t *testing.T) {
	// Initialize the email service with test credentials
	emailService := email.NewEmailService("smtp.example.com", "587", "test@example.com", "testpassword")
	err := emailService.SendEmailHTML("invalid-email", "Test HTML Subject", "<h1>Test HTML Body</h1>")
	assert.NotNil(t, err, "Expected an error but got nil")
}

// TestSendPasswordResetSuccess tests that sending a password reset email with valid config succeeds.
func TestSendPasswordResetSuccess(t *testing.T) {
	// Initialize the email service with test credentials
	config := config.LoadConfig("E:/workspace/auth/.env")
	emailService := email.NewEmailService(config.SMTPServer, config.SMTPPort, config.SMTPID, config.SMTPPassword)
	resetLink := fmt.Sprintf("https://yourdomain.com/reset-password?token=%s", "21312312312")
	err := emailService.SendPasswordReset("dgkwon90@naver.com", resetLink, 30)
	assert.Nil(t, err, "Expected no error but got one")
}
