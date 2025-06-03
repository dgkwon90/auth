package entity

import "time"

type PasswordResetTokenEntity struct {
	Id        int64     `db:"id" json:"id"`
	UserId    int64     `db:"user_id" json:"userId"`
	Token     string    `db:"token" json:"token"`
	ExpiredAt time.Time `db:"expired_at" json:"expiredAt"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
