package service

import (
	"auth/internal/dto"
	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/service/email"
	"auth/pkg/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type AuthService struct {
	dbPool       *pgxpool.Pool
	userRepo     repository.UserRepository
	profileRepo  repository.ProfileRepository
	jwtService   *JwtService
	emailService *email.EmailService
}

func NewAuthService(dbPool *pgxpool.Pool, userRepo repository.UserRepository, profileRepo repository.ProfileRepository, jwtService *JwtService, emailService *email.EmailService) *AuthService {
	return &AuthService{dbPool, userRepo, profileRepo, jwtService, emailService}
}

func (s *AuthService) RegisterUser(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 트랜잭션 시작
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("RegisterUser: begin tx failed", "error", err)
		return nil, err
	}
	defer tx.Rollback(ctx) // 성공하면 noop, 실패하면 롤백

	// 1. 이메일 중복 확인 - 트랜잭션 내에서 확인
	existingUser, err := s.userRepo.FindByEmailTx(ctx, tx, req.Email) // 트랜잭션 버전 필요
	if err != nil {
		slog.Error("RegisterUser: find email failed", "error", err)
		return nil, err
	}
	if existingUser != nil {
		slog.Warn("RegisterUser: email exists", "email", req.Email)
		return nil, errors.New("email already exists")
	}

	// 2. 비밀번호 해시
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		slog.Error("RegisterUser: hash password failed", "error", err)
		return nil, err
	}

	now := time.Now()
	// 3. UserEntity 생성
	userEntity := &entity.UserEntity{
		Email:        req.Email,
		PasswordHash: hashed,
		Provider:     "local", // 기본값 설정
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 트랜잭션으로 사용자 생성
	newUserId, err := s.userRepo.CreateTx(ctx, tx, userEntity)
	if err != nil {
		tx.Rollback(ctx)
		slog.Error("RegisterUser: create user failed, rollback", "error", err)
		return nil, err
	}

	// 4. ProfileEntity 생성 및 저장 (선택값 처리)
	profileEntity := &entity.ProfileEntity{
		UserId:      newUserId,
		Name:        req.Name,
		BirthDate:   parseDate(req.BirthDate),
		GenderCode:  entity.GenderCode(req.GenderCode),
		PhoneNumber: req.PhoneNumber,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// 트랜잭션으로 프로필 생성
	err = s.profileRepo.CreateTx(ctx, tx, profileEntity)
	if err != nil {
		slog.Error("RegisterUser: create profile failed", "error", err)
		return nil, err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(ctx); err != nil {
		slog.Error("RegisterUser: commit failed", "error", err)
		return nil, err
	}

	slog.Info("RegisterUser: success", "userId", newUserId, "email", req.Email)
	result := &dto.RegisterResponse{
		Email:       userEntity.Email,
		Name:        profileEntity.Name,                           // ProfileFromEntity가 반환하는 *Profile의 필드 사용
		BirthDate:   profileEntity.BirthDate.Format("2006-01-02"), // YYYY-MM-DD
		GenderCode:  string(profileEntity.GenderCode),
		PhoneNumber: profileEntity.PhoneNumber,
	}

	return result, nil
}

// 날짜 파싱 유틸
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func (s *AuthService) Login(ctx context.Context, cmd *dto.LoginRequest, deviceInfo string) (*dto.LoginResponse, error) {
	// 1. 이메일로 사용자 찾기
	u, err := s.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		slog.Error("Login: find by email failed", "error", err)
		return nil, err
	}
	if u == nil {
		slog.Warn("Login: user not found", "email", cmd.Email)
		return nil, errors.New("user not found")
	}

	// 2. 비밀번호 검증
	if !utils.CheckPasswordHash(cmd.Password, u.PasswordHash) {
		slog.Warn("Login: invalid password", "email", cmd.Email)
		return nil, errors.New("invalid password")
	}

	// 3. 기존 device의 refresh token 삭제 (동일 디바이스 중복 로그인 방지)
	err = s.userRepo.DeleteByUserIdAndDevice(ctx, u.Id, deviceInfo)
	if err != nil {
		slog.Error("Login: delete old refresh token failed", "userId", u.Id, "error", err)
		return nil, err
	}

	// 4. JWT 토큰 생성
	accessToken, err := s.jwtService.GenerateToken(u.Id)
	if err != nil {
		slog.Error("Login: generate access token failed", "userId", u.Id, "error", err)
		return nil, err
	}
	// 5. Refresh Token 생성 및 저장
	refreshToken, err := s.jwtService.GenerateRefreshToken(u.Id, deviceInfo)
	if err != nil {
		slog.Error("Login: generate refresh token failed", "userId", u.Id, "error", err)
		return nil, err
	}

	// refresh_tokens 테이블에 저장 (deviceInfo는 cmd.DeviceInfo에서 받아온다고 가정)
	rt := &entity.RefreshTokenEntity{
		UserId:     u.Id,
		Token:      refreshToken,
		DeviceInfo: deviceInfo,
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(7 * 24 * time.Hour),
	}

	err = s.userRepo.InsertRefreshToken(ctx, rt)
	if err != nil {
		slog.Error("Login: insert refresh token failed", "userId", u.Id, "error", err)
		return nil, err
	}

	// 6. 응답
	result := &dto.LoginResponse{
		UserId:       u.Id,
		Email:        u.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	slog.Info("Login: success", "userId", u.Id, "email", u.Email)
	return result, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	userId, deviceInfo, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		slog.Warn("RefreshToken: invalid refresh token", "error", err)
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}
	rtRecord, err := s.userRepo.FindByUserDeviceAndToken(ctx, userId, deviceInfo, refreshToken)
	if err != nil || rtRecord == nil {
		slog.Warn("RefreshToken: token not found", "userId", userId)
		return "", "", errors.New("refresh token not found")
	}
	if time.Now().After(rtRecord.ExpiredAt) {
		_ = s.userRepo.DeleteRefreshToken(ctx, userId, refreshToken)
		slog.Warn("RefreshToken: token expired", "userId", userId)
		return "", "", errors.New("refresh token expired")
	}
	// 기존 refresh token 삭제(재발급 시)
	_ = s.userRepo.DeleteRefreshToken(ctx, userId, refreshToken)
	// 새 refresh token 발급 및 저장
	newRefreshToken, err := s.jwtService.GenerateRefreshToken(userId, deviceInfo)
	if err != nil {
		slog.Error("RefreshToken: generate new refresh token failed", "userId", userId, "error", err)
		return "", "", err
	}
	rt := &entity.RefreshTokenEntity{
		UserId:     userId,
		Token:      newRefreshToken,
		DeviceInfo: deviceInfo,
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(7 * 24 * time.Hour),
	}
	err = s.userRepo.InsertRefreshToken(ctx, rt)
	if err != nil {
		slog.Error("RefreshToken: insert new refresh token failed", "userId", userId, "error", err)
		return "", "", err
	}
	accessToken, err := s.jwtService.GenerateToken(userId)
	if err != nil {
		slog.Error("RefreshToken: generate access token failed", "userId", userId, "error", err)
		return "", "", err
	}
	slog.Info("RefreshToken: success", "userId", userId)
	return accessToken, newRefreshToken, nil
}

// 이하 FindEmail, ForgotPassword, ResetPassword 등도 유사하게 구현
func (s *AuthService) FindEmail(ctx context.Context, cmd *dto.FindEmailRequest) (*dto.FindEmailResponse, error) {
	profile, err := s.profileRepo.FindByPhoneNumber(ctx, cmd.PhoneNumber)
	if err != nil || profile == nil {
		slog.Warn("FindEmail: profile not found", "phone", cmd.PhoneNumber)
		return nil, errors.New("not found")
	}

	user, err := s.userRepo.FindById(ctx, profile.UserId)
	if err != nil || user == nil {
		slog.Warn("FindEmail: user not found", "userId", profile.UserId)
		return nil, errors.New("not found")
	}

	slog.Info("FindEmail: success", "userId", user.Id)
	return &dto.FindEmailResponse{
		Email: utils.MaskEmail(user.Email),
	}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("ForgotPassword: begin tx failed", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		slog.Error("ForgotPassword: find user failed", "error", err)
		return err
	}
	if user == nil {
		slog.Warn("ForgotPassword: user not found", "email", email)
		return errors.New("user not found")
	}
	// 토큰 생성 (간단 예시, 실제로는 더 안전하게)
	token := utils.GenerateRandomString(32)
	expireMinutes := 30
	expiredAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)
	err = s.userRepo.SavePasswordResetToken(ctx, user.Id, token, expiredAt)
	if err != nil {
		slog.Error("ForgotPassword: save token failed", "error", err)
		return err
	}

	// 이메일 전송
	resetLink := fmt.Sprintf("https://yourdomain.com/reset-password?token=%s", token)
	err = s.emailService.SendPasswordReset(email, resetLink, expireMinutes)
	if err != nil {
		slog.Error("ForgotPassword: send email failed", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("ForgotPassword: commit failed", "error", err)
		return err
	}
	slog.Info("ForgotPassword: success", "userId", user.Id)
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("ResetPassword: begin tx failed", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	resetInfo, err := s.userRepo.FindByPasswordResetToken(ctx, token)
	if err != nil {
		slog.Error("ResetPassword: find token failed", "error", err)
		return err
	}
	if resetInfo == nil || time.Now().After(resetInfo.ExpiredAt) || resetInfo.Used {
		slog.Warn("ResetPassword: invalid, expired, or used token", "token", token)
		return errors.New("invalid, expired, or already used token")
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		slog.Error("ResetPassword: hash failed", "error", err)
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, resetInfo.UserId, hashed); err != nil {
		slog.Error("ResetPassword: update password failed", "error", err)
		return err
	}
	// used=true로 업데이트
	if err := s.userRepo.ExpirePasswordResetToken(ctx, token); err != nil {
		slog.Error("ResetPassword: expire token failed", "error", err)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		slog.Error("ResetPassword: commit failed", "error", err)
		return err
	}
	slog.Info("ResetPassword: success", "userId", resetInfo.UserId)
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userId int64) (*dto.ProfileResponse, error) {
	profile, err := s.profileRepo.FindByUserId(ctx, userId)
	if err != nil || profile == nil {
		return nil, errors.New("profile not found")
	}

	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	result := &dto.ProfileResponse{
		Email:       user.Email,
		Name:        profile.Name,
		BirthDate:   profile.BirthDate.Format("2006-01-02"), // YYYY-MM-DD
		GenderCode:  string(profile.GenderCode),
		PhoneNumber: profile.PhoneNumber,
	}
	return result, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userId int64, cmd *dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("UpdateProfile: begin tx failed", "error", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	profile, err := s.profileRepo.FindByUserId(ctx, userId)
	if err != nil {
		slog.Error("UpdateProfile: find profile failed", "error", err)
		return nil, err
	}
	if profile == nil {
		slog.Warn("UpdateProfile: profile not found", "userId", userId)
		return nil, errors.New("profile not found")
	}

	existing, err := s.profileRepo.FindByPhoneNumber(ctx, cmd.PhoneNumber)
	if err != nil {
		slog.Error("UpdateProfile: find by phone failed", "error", err)
		return nil, err
	}
	if existing != nil && existing.UserId != userId {
		slog.Warn("UpdateProfile: phone already used", "phone", cmd.PhoneNumber)
		return nil, errors.New("이미 사용 중인 전화번호입니다")
	}

	profile.Name = cmd.Name
	profile.BirthDate, err = time.Parse("2006-01-02", cmd.BirthDate)
	if err != nil {
		slog.Error("UpdateProfile: parse birthdate failed", "error", err)
		return nil, err
	}
	profile.GenderCode = entity.GenderCode(cmd.GenderCode)
	profile.PhoneNumber = cmd.PhoneNumber
	profile.UpdatedAt = time.Now()
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		slog.Error("UpdateProfile: update failed", "error", err)
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		slog.Error("UpdateProfile: commit failed", "error", err)
		return nil, err
	}
	slog.Info("UpdateProfile: success", "userId", userId)
	result := &dto.ProfileResponse{
		Name:        profile.Name,
		BirthDate:   profile.BirthDate.Format("2006-01-02"),
		GenderCode:  string(profile.GenderCode),
		PhoneNumber: profile.PhoneNumber,
	}
	return result, nil
}

func (s *AuthService) JwtSvc() *JwtService {
	return s.jwtService
}

func (s *AuthService) Logout(ctx context.Context, userId int64, refreshToken, deviceInfo string) error {
	return s.userRepo.DeleteRefreshToken(ctx, userId, refreshToken)
}

func (s *AuthService) ChangePassword(ctx context.Context, userId int64, currentPassword, newPassword string) error {
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("ChangePassword: begin tx failed", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil {
		slog.Error("ChangePassword: find user failed", "error", err)
		return err
	}
	if user == nil {
		slog.Warn("ChangePassword: user not found", "userId", userId)
		return errors.New("user not found")
	}
	if !utils.CheckPasswordHash(currentPassword, user.PasswordHash) {
		slog.Warn("ChangePassword: current password incorrect", "userId", userId)
		return errors.New("current password is incorrect")
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		slog.Error("ChangePassword: hash failed", "error", err)
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, userId, hashed); err != nil {
		slog.Error("ChangePassword: update password failed", "error", err)
		return err
	}
	_ = s.userRepo.DeleteAllRefreshTokens(ctx, userId)
	if err := tx.Commit(ctx); err != nil {
		slog.Error("ChangePassword: commit failed", "error", err)
		return err
	}
	slog.Info("ChangePassword: success", "userId", userId)
	return nil
}

func (s *AuthService) CheckPassword(ctx context.Context, userId int64, password string) error {
	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}
	return nil
}

func (s *AuthService) DeleteProfile(ctx context.Context, userId int64) error {
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		slog.Error("DeleteProfile: begin tx failed", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	if err := s.userRepo.Delete(ctx, userId); err != nil {
		slog.Error("DeleteProfile: user soft delete failed", "error", err)
		return err
	}
	_ = s.userRepo.DeleteAllRefreshTokens(ctx, userId)
	if err := tx.Commit(ctx); err != nil {
		slog.Error("DeleteProfile: commit failed", "error", err)
		return err
	}
	slog.Info("DeleteProfile: success", "userId", userId)
	return nil
}
