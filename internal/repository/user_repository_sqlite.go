package repository

import (
	"auth/internal/entity"
	"context"
	"log"
	"time"

	"zombiezen.com/go/sqlite"
)

type userRepositorySqlite struct {
	db *sqlite.Conn
}

// NewUserRepositorySqlite returns a new sqlite-based UserRepository.
func NewUserRepositorySqlite(conn *sqlite.Conn) UserRepository {
	repo := &userRepositorySqlite{db: conn}
	if err := repo.createTable(context.Background()); err != nil {
		log.Printf("[sqlite] Error creating tables: %v", err)
	}
	return repo
}

func (r *userRepositorySqlite) createTable(_ context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			provider TEXT DEFAULT 'local',
			provider_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			birth_date TEXT,
			gender_code TEXT,
			phone_number TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			device_info TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expired_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expired_at DATETIME,
			used BOOLEAN DEFAULT 0,
			UNIQUE (user_id)
		);`,
	}
	for _, q := range stmts {
		stmt, err := r.db.Prepare(q)
		if err != nil {
			return err
		}
		_, err = stmt.Step()
		err2 := stmt.Finalize()
		if err != nil {
			return err
		}
		if err2 != nil {
			return err2
		}
	}
	return nil
}

// CreateTx creates a user in sqlite (no real tx used)
func (r *userRepositorySqlite) CreateTx(_ context.Context, _ interface{}, user *entity.UserEntity) (int64, error) {
	stmt, err := r.db.Prepare("INSERT INTO users (email, password_hash, provider, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)")
	if err != nil {
		return 0, err
	}
	stmt.BindText(1, user.Email)
	stmt.BindText(2, user.PasswordHash)
	stmt.BindText(3, user.Provider)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return 0, err
	}
	if err2 != nil {
		return 0, err2
	}
	id := r.db.LastInsertRowID()
	return id, nil
}

// FindByID returns a user by ID.
func (r *userRepositorySqlite) FindByID(_ context.Context, id int64) (*entity.UserEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, email, password_hash, provider, provider_id, created_at, updated_at, deleted_at FROM users WHERE id = ? AND deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	stmt.BindInt64(1, id)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var u entity.UserEntity
	u.ID = stmt.ColumnInt64(0)
	u.Email = stmt.ColumnText(1)
	u.PasswordHash = stmt.ColumnText(2)
	u.Provider = stmt.ColumnText(3)
	providerID := stmt.ColumnText(4)
	if providerID != "" {
		u.ProviderID = &providerID
	}
	_ = stmt.Finalize()
	return &u, nil
}

// FindByEmail returns a user by email.
func (r *userRepositorySqlite) FindByEmail(_ context.Context, email string) (*entity.UserEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, email, password_hash, provider, provider_id, created_at, updated_at, deleted_at FROM users WHERE email = ? AND deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	stmt.BindText(1, email)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var u entity.UserEntity
	u.ID = stmt.ColumnInt64(0)
	u.Email = stmt.ColumnText(1)
	u.PasswordHash = stmt.ColumnText(2)
	u.Provider = stmt.ColumnText(3)
	providerID := stmt.ColumnText(4)
	if providerID != "" {
		u.ProviderID = &providerID
	}
	_ = stmt.Finalize()
	return &u, nil
}

// FindByEmailTx returns a user by email in a tx (sqlite ignores tx).
func (r *userRepositorySqlite) FindByEmailTx(_ context.Context, _ interface{}, email string) (*entity.UserEntity, error) {
	return r.FindByEmail(context.Background(), email)
}

// UpdatePassword updates a user's password hash.
func (r *userRepositorySqlite) UpdatePassword(_ context.Context, id int64, passwordHash string) error {
	stmt, err := r.db.Prepare("UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		return err
	}
	stmt.BindText(1, passwordHash)
	stmt.BindInt64(2, id)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// Delete soft-deletes a user.
func (r *userRepositorySqlite) Delete(_ context.Context, id int64) error {
	stmt, err := r.db.Prepare("UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, id)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// InsertRefreshToken inserts a refresh token.
func (r *userRepositorySqlite) InsertRefreshToken(_ context.Context, rt *entity.RefreshTokenEntity) error {
	stmt, err := r.db.Prepare("INSERT INTO refresh_tokens (user_id, token, device_info, created_at, expired_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?)")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, rt.UserID)
	stmt.BindText(2, rt.Token)
	stmt.BindText(3, rt.DeviceInfo)
	if rt.ExpiredAt.IsZero() {
		stmt.BindNull(4)
	} else {
		stmt.BindText(4, rt.ExpiredAt.Format("2006-01-02 15:04:05"))
	}
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// DeleteByUserIDAndDevice deletes a refresh token by user and device.
func (r *userRepositorySqlite) DeleteByUserIDAndDevice(_ context.Context, userID int64, deviceInfo string) error {
	stmt, err := r.db.Prepare("DELETE FROM refresh_tokens WHERE user_id = ? AND device_info = ?")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, userID)
	stmt.BindText(2, deviceInfo)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// FindRefreshToken finds a refresh token by token string.
func (r *userRepositorySqlite) FindRefreshToken(_ context.Context, token string) (*entity.RefreshTokenEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, token, device_info, created_at, expired_at FROM refresh_tokens WHERE token = ?")
	if err != nil {
		return nil, err
	}
	stmt.BindText(1, token)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var rt entity.RefreshTokenEntity
	rt.ID = stmt.ColumnInt64(0)
	rt.UserID = stmt.ColumnInt64(1)
	rt.Token = stmt.ColumnText(2)
	rt.DeviceInfo = stmt.ColumnText(3)
	// created_at, expired_at 파싱
	expiredAtStr := stmt.ColumnText(5)
	if expiredAtStr != "" {
		t, err := time.Parse("2006-01-02 15:04:05", expiredAtStr)
		if err == nil {
			rt.ExpiredAt = t
		}
	}
	err2 := stmt.Finalize()
	if err2 != nil {
		log.Printf("stmt.Finalize error: %v", err2)
	}
	return &rt, nil
}

func (r *userRepositorySqlite) FindByUserDeviceAndToken(_ context.Context, userID int64, deviceInfo, token string) (*entity.RefreshTokenEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, token, device_info, created_at, expired_at FROM refresh_tokens WHERE user_id = ? AND device_info = ? AND token = ?")
	if err != nil {
		return nil, err
	}
	stmt.BindInt64(1, userID)
	stmt.BindText(2, deviceInfo)
	stmt.BindText(3, token)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var rt entity.RefreshTokenEntity
	rt.ID = stmt.ColumnInt64(0)
	rt.UserID = stmt.ColumnInt64(1)
	rt.Token = stmt.ColumnText(2)
	rt.DeviceInfo = stmt.ColumnText(3)
	expiredAtStr := stmt.ColumnText(5)
	if expiredAtStr != "" {
		t, err := time.Parse("2006-01-02 15:04:05", expiredAtStr)
		if err == nil {
			rt.ExpiredAt = t
		}
	}
	if err2 := stmt.Finalize(); err2 != nil {
		log.Printf("stmt.Finalize error: %v", err2)
	}
	return &rt, nil
}

func (r *userRepositorySqlite) DeleteRefreshToken(_ context.Context, userID int64, token string) error {
	stmt, err := r.db.Prepare("DELETE FROM refresh_tokens WHERE user_id = ? AND token = ?")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, userID)
	stmt.BindText(2, token)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

func (r *userRepositorySqlite) DeleteAllRefreshTokens(_ context.Context, userID int64) error {
	stmt, err := r.db.Prepare("DELETE FROM refresh_tokens WHERE user_id = ?")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, userID)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

func (r *userRepositorySqlite) SavePasswordResetToken(_ context.Context, userID int64, token string, expiredAt time.Time) error {
	stmt, err := r.db.Prepare("INSERT OR REPLACE INTO password_reset_tokens (user_id, token, expired_at, used) VALUES (?, ?, ?, 0)")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, userID)
	stmt.BindText(2, token)
	stmt.BindText(3, expiredAt.Format("2006-01-02 15:04:05"))
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// FindByPasswordResetToken finds a password reset token entity by token string (SQLite).
func (r *userRepositorySqlite) FindByPasswordResetToken(_ context.Context, token string) (*entity.PasswordResetTokenEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, token, expired_at, used FROM password_reset_tokens WHERE token = ?")
	if err != nil {
		return nil, err
	}
	stmt.BindText(1, token)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var prt entity.PasswordResetTokenEntity
	prt.ID = stmt.ColumnInt64(0)
	prt.UserID = stmt.ColumnInt64(1)
	prt.Token = stmt.ColumnText(2)
	expiredAtStr := stmt.ColumnText(3)
	t, parseErr := time.Parse("2006-01-02 15:04:05", expiredAtStr)
	if parseErr != nil {
		_ = stmt.Finalize()
		return nil, parseErr
	}
	prt.ExpiredAt = t
	prt.Used = stmt.ColumnInt64(4) != 0
	if err2 := stmt.Finalize(); err2 != nil {
		log.Printf("stmt.Finalize error: %v", err2)
	}
	return &prt, nil
}

func (r *userRepositorySqlite) ExpirePasswordResetToken(_ context.Context, token string) error {
	stmt, err := r.db.Prepare("UPDATE password_reset_tokens SET used = 1 WHERE token = ?")
	if err != nil {
		return err
	}
	stmt.BindText(1, token)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}
