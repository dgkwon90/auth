// Package entity provides database entity definitions for the authentication service.
package entity

import "time"

// RefreshTokenEntity represents a refresh token record.
type RefreshTokenEntity struct {
	ID         int64     `db:"id" json:"id"`
	UserID     int64     `db:"user_id" json:"userID"`
	Token      string    `db:"token" json:"token"`
	DeviceInfo string    `db:"device_info" json:"deviceInfo"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	ExpiredAt  time.Time `db:"expired_at" json:"expiredAt"`
}
