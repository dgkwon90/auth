// Package repository provides database access and persistence logic.
package repository

import (
	"auth/internal/entity"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// UserRepository defines user-related database operations.
type UserRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, user *entity.UserEntity) (int64, error)
	FindByID(ctx context.Context, id int64) (*entity.UserEntity, error)
	FindByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	FindByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*entity.UserEntity, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	Delete(ctx context.Context, id int64) error
	InsertRefreshToken(ctx context.Context, rt *entity.RefreshTokenEntity) error
	DeleteByUserIDAndDevice(ctx context.Context, userID int64, deviceInfo string) error
	FindRefreshToken(ctx context.Context, token string) (*entity.RefreshTokenEntity, error)
	FindByUserDeviceAndToken(ctx context.Context, userID int64, deviceInfo, token string) (*entity.RefreshTokenEntity, error)
	DeleteRefreshToken(ctx context.Context, userID int64, token string) error
	DeleteAllRefreshTokens(ctx context.Context, userID int64) error
	SavePasswordResetToken(ctx context.Context, userID int64, token string, expiredAt time.Time) error
	FindByPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetTokenEntity, error)
	ExpirePasswordResetToken(ctx context.Context, token string) error
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(dbPool *pgxpool.Pool) UserRepository {
	repo := &userRepository{dbPool: dbPool}
	if err := repo.createTable(context.Background()); err != nil {
		slog.Warn("Error creating tables", "error", err)
	}
	return repo
}

// userRepository implements UserRepository interface.
type userRepository struct {
	dbPool *pgxpool.Pool
}

// createTable: users·refresh_tokens 테이블 생성
func (r *userRepository) createTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
		provider VARCHAR(50) DEFAULT 'local',
        provider_id VARCHAR(255),
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW(),
        deleted_at TIMESTAMPTZ
    );
    CREATE TABLE IF NOT EXISTS refresh_tokens (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token VARCHAR(512) NOT NULL,
        device_info VARCHAR(255),
        created_at TIMESTAMPTZ DEFAULT NOW(),
        expired_at TIMESTAMPTZ
    );
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(512) NOT NULL,
		expired_at TIMESTAMPTZ,
		used BOOLEAN DEFAULT false,
		UNIQUE (user_id)
	);`
	_, err := r.dbPool.Exec(ctx, query)
	return err
}

// func (r *UserRepository) Pool() *pgxpool.Pool {
// 	return r.dbPool
// }

// CreateTx: 트랜잭션(tx, 티엑스)으로 users 생성
func (r *userRepository) CreateTx(ctx context.Context, tx pgx.Tx, user *entity.UserEntity) (int64, error) {
	var id int64
	query := `INSERT INTO users (email, password_hash, provider, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`
	err := tx.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Provider, user.CreatedAt, user.UpdatedAt,
	).Scan(&id)
	return id, err
}

// Create: users 단일 생성
// func (r *UserRepository) Create(ctx context.Context, user *entity.UserEntity) error {
// 	query := `
//         INSERT INTO users (email, password_hash, created_at, updated_at)
//         VALUES ($1, $2, $3, $4)
//         RETURNING id
//     `
// 	return r.dbPool.QueryRow(ctx, query,
// 		user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
// 	).Scan(&user.Id)
// }

// FindById: ID로 사용자 조회
func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.UserEntity, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL`
	u := &entity.UserEntity{}
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// FindByEmail: 이메일로 사용자 조회
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	query := `SELECT id, email, password_hash, provider, provider_id, created_at, updated_at, deleted_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL`
	u := &entity.UserEntity{}
	err := r.dbPool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Provider, &u.ProviderID,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// FindByEmailTx: 트랜잭션 내에서 이메일로 사용자 조회
func (r *userRepository) FindByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*entity.UserEntity, error) {
	query := `SELECT id, email, password_hash,  provider, provider_id, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`
	u := &entity.UserEntity{}
	err := tx.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Provider, &u.ProviderID,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// UpdatePassword: 비밀번호(hash, 해시) 변경
func (r *userRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `UPDATE users
        SET password_hash = $1, updated_at = NOW()
        WHERE id = $2`
	_, err := r.dbPool.Exec(ctx, query, passwordHash, id)
	return err
}

// Delete: soft delete (deleted_at 셋)
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users
        SET deleted_at = NOW()
        WHERE id = $1`
	_, err := r.dbPool.Exec(ctx, query, id)
	return err
}

// InsertRefreshToken: 리프레시 토큰 등록
func (r *userRepository) InsertRefreshToken(ctx context.Context, rt *entity.RefreshTokenEntity) error {
	query := `INSERT INTO refresh_tokens (user_id, token, device_info, created_at, expired_at)
        VALUES ($1, $2, $3, NOW(), $4)`
	_, err := r.dbPool.Exec(ctx, query, rt.UserID, rt.Token, rt.DeviceInfo, rt.ExpiredAt)
	return err
}

// DeleteByUserIdAndDevice: 특정 디바이스 토큰 삭제
func (r *userRepository) DeleteByUserIDAndDevice(ctx context.Context, userID int64, deviceInfo string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id=$1 AND device_info=$2`
	_, err := r.dbPool.Exec(ctx, query, userID, deviceInfo)
	return err
}

// FindRefreshToken: 토큰 값 단일 조회
func (r *userRepository) FindRefreshToken(ctx context.Context, token string) (*entity.RefreshTokenEntity, error) {
	query := `SELECT id, user_id, token, device_info, created_at, expired_at
        FROM refresh_tokens
        WHERE token = $1`
	rt := &entity.RefreshTokenEntity{}
	err := r.dbPool.QueryRow(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.DeviceInfo, &rt.CreatedAt, &rt.ExpiredAt,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

// FindByUserDeviceAndToken: user+device+token 으로 조회
func (r *userRepository) FindByUserDeviceAndToken(ctx context.Context, userID int64, deviceInfo, token string) (*entity.RefreshTokenEntity, error) {
	query := `SELECT id, user_id, token, device_info, created_at, expired_at
        FROM refresh_tokens
        WHERE user_id=$1 AND device_info=$2 AND token=$3`
	rt := &entity.RefreshTokenEntity{}
	err := r.dbPool.QueryRow(ctx, query, userID, deviceInfo, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.DeviceInfo, &rt.CreatedAt, &rt.ExpiredAt,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

// DeleteRefreshToken: 로그아웃용 단일 토큰 삭제
func (r *userRepository) DeleteRefreshToken(ctx context.Context, userID int64, token string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1 AND token = $2`
	_, err := r.dbPool.Exec(ctx, query, userID, token)
	return err
}

// DeleteAllRefreshTokens: 회원탈퇴·강제 로그아웃용 전체 삭제
func (r *userRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.dbPool.Exec(ctx, query, userID)
	return err
}

func (r *userRepository) SavePasswordResetToken(ctx context.Context, userID int64, token string, expiredAt time.Time) error {
	_, err := r.dbPool.Exec(ctx, `INSERT INTO password_reset_tokens (user_id, token, expired_at, used)
         VALUES ($1, $2, $3, false)
         ON CONFLICT (user_id) DO UPDATE SET token = $2, expired_at = $3, used = false`,
		userID, token, expiredAt)
	return err
}

func (r *userRepository) FindByPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetTokenEntity, error) {
	row := r.dbPool.QueryRow(ctx, `SELECT user_id, token, expired_at, used FROM password_reset_tokens WHERE token=$1 AND used=false`, token)
	var info entity.PasswordResetTokenEntity
	err := row.Scan(&info.UserID, &info.Token, &info.ExpiredAt, &info.Used)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *userRepository) ExpirePasswordResetToken(ctx context.Context, token string) error {
	_, err := r.dbPool.Exec(ctx, `UPDATE password_reset_tokens SET used=true WHERE token=$1`, token)
	return err
}
