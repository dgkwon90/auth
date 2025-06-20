// Package dto provides data transfer objects for API requests and responses in the authentication service.
package dto

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	Name        string `json:"name" validate:"required,namekr"`
	BirthDate   string `json:"birthDate" validate:"required,len=10"`
	GenderCode  string `json:"genderCode" validate:"required,oneof=M F O N U"`
	PhoneNumber string `json:"phoneNumber" validate:"required,phonekr"`
}

// RegisterResponse represents a user registration response.
type RegisterResponse struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	BirthDate   string `json:"birthDate"`
	GenderCode  string `json:"genderCode"`
	PhoneNumber string `json:"phoneNumber"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// LoginResponse represents a user login response.
type LoginResponse struct {
	UserID       int64  `json:"userId"`
	Email        string `json:"email"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenRequest represents a refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// RefreshTokenResponse represents a refresh token response.
type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// FindEmailRequest represents a request to find an email by phone number.
type FindEmailRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,phonekr"`
}

// FindEmailResponse represents a response containing an email.
type FindEmailResponse struct {
	Email string `json:"email"`
}

// ForgotPasswordRequest represents a request to send a password reset email.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents a request to reset a password.
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=128"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
