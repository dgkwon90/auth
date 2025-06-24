// Package repository provides database access and persistence logic.
package repository

import (
	"context"
	"errors"

	"log/slog"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"zombiezen.com/go/sqlite"

	"auth/internal/entity"
)

// ProfileRepository defines profile-related database operations.
type ProfileRepository interface {
	CreateTx(ctx context.Context, tx interface{}, p *entity.ProfileEntity) error
	FindByUserID(ctx context.Context, userID int64) (*entity.ProfileEntity, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.ProfileEntity, error)
	Update(ctx context.Context, p *entity.ProfileEntity) error
	CreateTable(ctx context.Context) error
}

type profileRepository struct {
	dbPool *pgxpool.Pool
}

// NewProfileRepository creates a new ProfileRepository instance.
func NewProfileRepository(dbPool *pgxpool.Pool) ProfileRepository {
	r := &profileRepository{dbPool: dbPool}
	if err := r.CreateTable(context.Background()); err != nil {
		slog.Warn("Error creating profiles table", "error", err)
	}
	return r
}

// NewProfileRepositoryAuto returns a ProfileRepository for the given DB type.
func NewProfileRepositoryAuto(dbType string, pgxPool *pgxpool.Pool, sqliteConn interface{}) ProfileRepository {
	switch dbType {
	case "sqlite":
		if conn, ok := sqliteConn.(*sqlite.Conn); ok {
			return NewProfileRepositorySqlite(conn)
		}
		panic("sqliteConn is not *sqlite.Conn")
	case "postgres":
		fallthrough
	default:
		return NewProfileRepository(pgxPool)
	}
}

// CreateTable creates the profiles table if it does not exist
func (r *profileRepository) CreateTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS gender_codes (
		code CHAR(1) PRIMARY KEY,          -- M, F, O, N, U
		description VARCHAR(50) NOT NULL   -- 예: 'male'
	);

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM gender_codes) THEN
			INSERT INTO gender_codes(code, description) VALUES
				('M','male'),
				('F','female'),
				('O','other '),
				('N','non_binary'),
				('U','unspecified');
		END IF;
	END $$;

	CREATE TABLE IF NOT EXISTS profiles (
		id           SERIAL PRIMARY KEY,
		user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name         VARCHAR(255) NOT NULL,
		birth_date   DATE,
		gender_code  CHAR(1) NOT NULL REFERENCES gender_codes(code),
		phone_number VARCHAR(16) NOT NULL UNIQUE, 
		created_at   TIMESTAMPTZ DEFAULT NOW(),
		updated_at   TIMESTAMPTZ DEFAULT NOW()
	);
	`
	_, err := r.dbPool.Exec(ctx, query)
	return err
}

// CreateTx inserts a profile within a transaction
func (r *profileRepository) CreateTx(ctx context.Context, tx interface{}, p *entity.ProfileEntity) error {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return errors.New("tx is not pgx.Tx")
	}
	query := `INSERT INTO profiles (user_id, name, birth_date, gender_code, phone_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	return pgxTx.QueryRow(ctx, query,
		p.UserID, p.Name, p.BirthDate, p.GenderCode, p.PhoneNumber, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
}

// Create adds a new profile record
// func (r *profileRepository) Create(ctx context.Context, p *entity.ProfileEntity) error {
// 	query := `
//         INSERT INTO profiles (user_id, name, birth_date, gender_code, phone_number, created_at, updated_at)
//         VALUES ($1, $2, $3, $4, $5, $6, $7)
//         RETURNING id
//     `
// 	return r.dbPool.QueryRow(ctx, query,
// 		p.UserId, p.Name, p.BirthDate, p.GenderCode, p.PhoneNumber, p.CreatedAt, p.UpdatedAt,
// 	).Scan(&p.Id)
// }

// FindByUserID retrieves a profile by user ID
func (r *profileRepository) FindByUserID(ctx context.Context, userID int64) (*entity.ProfileEntity, error) {
	query := `SELECT
		id,                 -- int64
        user_id,            -- int64
        name,               -- string
        birth_date,         -- time.Time
        gender_code,        -- string
        phone_number,       -- string
        created_at,         -- time.Time
        updated_at          -- time.Time
    FROM profiles
    WHERE user_id = $1`
	p := &entity.ProfileEntity{}
	err := r.dbPool.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.BirthDate, &p.GenderCode, &p.PhoneNumber, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// FindByPhoneNumber retrieves a profile by phone number
func (r *profileRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.ProfileEntity, error) {
	query := `
        SELECT
		    id,                 -- int64
            user_id,            -- int64
            name,               -- string
            birth_date,         -- time.Time
            gender_code,        -- string
            phone_number,       -- string
            created_at,         -- time.Time
            updated_at          -- time.Time
        FROM profiles
        WHERE phone_number = $1
    `
	p := &entity.ProfileEntity{}
	err := r.dbPool.QueryRow(ctx, query, phoneNumber).Scan(
		&p.ID, &p.UserID, &p.Name, &p.BirthDate, &p.GenderCode, &p.PhoneNumber, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// Update modifies an existing profile
func (r *profileRepository) Update(ctx context.Context, p *entity.ProfileEntity) error {
	query := `UPDATE profiles
        SET name = $1, birth_date = $2, gender_code = $3, phone_number = $4, updated_at = $5
        WHERE user_id = $6`
	cmd, err := r.dbPool.Exec(ctx, query,
		p.Name, p.BirthDate, p.GenderCode, p.PhoneNumber, p.UpdatedAt, p.UserID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("no profile updated")
	}
	return nil
}

// UpdateTx modifies an existing profile within a transaction
func (r *profileRepository) UpdateTx(ctx context.Context, tx pgx.Tx, p *entity.ProfileEntity) error {
	query := `UPDATE profiles
        SET name = $1, birth_date = $2, gender_code = $3, phone_number = $4, updated_at = $5
        WHERE user_id = $6`
	cmd, err := tx.Exec(ctx, query,
		p.Name, p.BirthDate, p.GenderCode, p.PhoneNumber, p.UpdatedAt, p.UserID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("no profile updated")
	}
	return nil
}
