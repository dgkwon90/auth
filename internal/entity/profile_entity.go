package entity

import "time"

type ProfileEntity struct {
	Id          int64      `db:"id" json:"id"`
	UserId      int64      `db:"user_id" json:"userId"`
	Name        string     `db:"name" json:"name"`
	BirthDate   time.Time  `db:"birth_date" json:"birthDate"`     // YYYY-MM-DD(ISO 8601)
	GenderCode  GenderCode `db:"gender_code" json:"genderCode"`   // 'M','F','O','N','U'
	PhoneNumber string     `db:"phone_number" json:"phoneNumber"` // E.164 국제표준
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
}

type GenderCode string

const (
	GenderCodeMale        GenderCode = "M" // male
	GenderCodeFemale      GenderCode = "F" // female
	GenderCodeOther       GenderCode = "O" // other
	GenderCodeNonBinary   GenderCode = "N" // non_binary
	GenderCodeUnspecified GenderCode = "U" // unspecified
)
