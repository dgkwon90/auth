// Package dto provides data transfer objects for user-related API requests and responses.
package dto

// ProfileResponse represents a user profile response.
type ProfileResponse struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	BirthDate   string `json:"birthDate"`
	GenderCode  string `json:"genderCode"`
	PhoneNumber string `json:"phoneNumber"`
}

// UpdateProfileRequest represents a request to update a user profile.
type UpdateProfileRequest struct {
	// Id int64 `json:"id"`
	// UserId      int64  `json:"userId"`
	Name        string `json:"name" validate:"required,namekr"`
	BirthDate   string `json:"birthDate" validate:"required,len=10"`
	GenderCode  string `json:"genderCode" validate:"required,oneof=M F O N U"`
	PhoneNumber string `json:"phoneNumber" validate:"required,phonekr"`
}

// ChangePasswordRequest represents a request to change a user's password.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required,min=8,max=128"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=128"`
}

// DeleteProfileRequest represents a request to delete a user profile.
type DeleteProfileRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
}
