package repository

import (
	"auth/internal/entity"
	"context"
	"log/slog"
	"time"

	"zombiezen.com/go/sqlite"
)

type profileRepositorySqlite struct {
	db *sqlite.Conn
}

// NewProfileRepositorySqlite returns a new sqlite-based ProfileRepository.
func NewProfileRepositorySqlite(conn *sqlite.Conn) ProfileRepository {
	r := &profileRepositorySqlite{db: conn}
	if err := r.createTable(context.Background()); err != nil {
		slog.Warn("[sqlite] Error creating profiles table", "error", err)
	}
	return r
}

func (r *profileRepositorySqlite) createTable(_ context.Context) error {
	stmt, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		birth_date TEXT,
		gender_code TEXT,
		phone_number TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return err
	}
	err = stmt.Finalize()
	return err
}

// CreateTx creates a profile in sqlite (no real tx used)
func (r *profileRepositorySqlite) CreateTx(_ context.Context, _ interface{}, p *entity.ProfileEntity) error {
	stmt, err := r.db.Prepare("INSERT INTO profiles (user_id, name, birth_date, gender_code, phone_number, created_at, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	stmt.BindInt64(1, p.UserID)
	stmt.BindText(2, p.Name)
	stmt.BindText(3, p.BirthDate.Format("2006-01-02"))
	stmt.BindText(4, string(p.GenderCode))
	stmt.BindText(5, p.PhoneNumber)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// FindByUserID returns a profile by user ID.
func (r *profileRepositorySqlite) FindByUserID(_ context.Context, userID int64) (*entity.ProfileEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, name, birth_date, gender_code, phone_number, created_at, updated_at FROM profiles WHERE user_id = ?")
	if err != nil {
		return nil, err
	}
	stmt.BindInt64(1, userID)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var p entity.ProfileEntity
	p.ID = stmt.ColumnInt64(0)
	p.UserID = stmt.ColumnInt64(1)
	p.Name = stmt.ColumnText(2)
	birthDateStr := stmt.ColumnText(3)
	if birthDateStr != "" {
		t, err := time.Parse("2006-01-02", birthDateStr)
		if err == nil {
			p.BirthDate = t
		}
	}
	p.GenderCode = entity.GenderCode(stmt.ColumnText(4))
	p.PhoneNumber = stmt.ColumnText(5)
	// created_at, updated_at 생략 가능
	_ = stmt.Finalize()
	return &p, nil
}

// FindByPhoneNumber returns a profile by phone number.
func (r *profileRepositorySqlite) FindByPhoneNumber(_ context.Context, phoneNumber string) (*entity.ProfileEntity, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, name, birth_date, gender_code, phone_number, created_at, updated_at FROM profiles WHERE phone_number = ?")
	if err != nil {
		return nil, err
	}
	stmt.BindText(1, phoneNumber)
	hasRow, err := stmt.Step()
	if err != nil {
		_ = stmt.Finalize()
		return nil, err
	}
	if !hasRow {
		_ = stmt.Finalize()
		return nil, nil
	}
	var p entity.ProfileEntity
	p.ID = stmt.ColumnInt64(0)
	p.UserID = stmt.ColumnInt64(1)
	p.Name = stmt.ColumnText(2)
	birthDateStr := stmt.ColumnText(3)
	if birthDateStr != "" {
		t, err := time.Parse("2006-01-02", birthDateStr)
		if err == nil {
			p.BirthDate = t
		}
	}
	p.GenderCode = entity.GenderCode(stmt.ColumnText(4))
	p.PhoneNumber = stmt.ColumnText(5)
	// created_at, updated_at 생략 가능
	_ = stmt.Finalize()
	return &p, nil
}

// Update updates a profile in sqlite.
func (r *profileRepositorySqlite) Update(_ context.Context, p *entity.ProfileEntity) error {
	stmt, err := r.db.Prepare("UPDATE profiles SET name = ?, birth_date = ?, gender_code = ?, phone_number = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?")
	if err != nil {
		return err
	}
	stmt.BindText(1, p.Name)
	stmt.BindText(2, p.BirthDate.Format("2006-01-02"))
	stmt.BindText(3, string(p.GenderCode))
	stmt.BindText(4, p.PhoneNumber)
	stmt.BindInt64(5, p.UserID)
	_, err = stmt.Step()
	err2 := stmt.Finalize()
	if err != nil {
		return err
	}
	return err2
}

// CreateTable creates the profiles table if it does not exist.
func (r *profileRepositorySqlite) CreateTable(ctx context.Context) error {
	return r.createTable(ctx)
}
