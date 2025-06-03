package entity

import "time"

type UserEntity struct {
	Id           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Provider     string     `db:"provider" json:"provider"` // defautl "local"
	ProviderId   *string    `db:"provider_id" json:"providerId"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}
