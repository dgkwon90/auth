// Package entity provides database entity definitions for the authentication service.
package entity

import "time"

// UserEntity represents a user record in the database.
type UserEntity struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Provider     string     `db:"provider" json:"provider"` // default "local"
	ProviderID   *string    `db:"provider_id" json:"providerId"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}
