// Package entity provides database entity definitions for the authentication service.
package entity

import "time"

// ProfileEntity represents a user profile record in the database.
type ProfileEntity struct {
	ID          int64      `db:"id" json:"id"`
	UserID      int64      `db:"user_id" json:"userID"`
	Name        string     `db:"name" json:"name"`
	BirthDate   time.Time  `db:"birth_date" json:"birthDate"`     // YYYY-MM-DD(ISO 8601)
	GenderCode  GenderCode `db:"gender_code" json:"genderCode"`   // 'M','F','O','N','U'
	PhoneNumber string     `db:"phone_number" json:"phoneNumber"` // E.164 국제표준
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
}

// GenderCode represents the gender code for a profile.
type GenderCode string

const (
	// GenderCodeMale is the constant for male gender code.
	GenderCodeMale GenderCode = "M" // male
	// GenderCodeFemale is the constant for female gender code.
	GenderCodeFemale GenderCode = "F" // female
	// GenderCodeOther is the constant for other gender code.
	GenderCodeOther GenderCode = "O" // other
	// GenderCodeNonBinary is the constant for non-b gender code.
	GenderCodeNonBinary GenderCode = "N" // non_binary gender code.
	// GenderCodeUnspecified is the constant for unspecified  gender code.
	GenderCodeUnspecified GenderCode = "U" // unspecified
)
