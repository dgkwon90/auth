package dto

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	Name        string `json:"name" validate:"required,namekr"`
	BirthDate   string `json:"birthDate" validate:"required,len=10"`
	GenderCode  string `json:"genderCode" validate:"required,oneof=M F O N U"`
	PhoneNumber string `json:"phoneNumber" validate:"required,phonekr"`
}

type RegisterResponse struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	BirthDate   string `json:"birthDate"`
	GenderCode  string `json:"genderCode"`
	PhoneNumber string `json:"phoneNumber"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginResponse struct {
	UserId       int64  `json:"userId"`
	Email        string `json:"email"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type FindEmailRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,phonekr"`
}

type FindEmailResponse struct {
	Email string `json:"email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=128"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
