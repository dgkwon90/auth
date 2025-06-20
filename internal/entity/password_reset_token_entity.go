// Package entity provides database entity definitions for the authentication service.
package entity

import "time"

// PasswordResetTokenEntity represents a password reset token record.
type PasswordResetTokenEntity struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"userID"`
	Token     string    `db:"token" json:"token"`
	ExpiredAt time.Time `db:"expired_at" json:"expiredAt"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
