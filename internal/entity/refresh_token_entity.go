package entity

import "time"

type RefreshTokenEntity struct {
	Id         int64     `db:"id" json:"id"`
	UserId     int64     `db:"user_id" json:"userId"`
	Token      string    `db:"token" json:"token"`
	DeviceInfo string    `db:"device_info" json:"deviceInfo"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	ExpiredAt  time.Time `db:"expired_at" json:"expiredAt"`
}
