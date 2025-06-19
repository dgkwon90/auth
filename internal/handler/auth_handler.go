// Package handler provides HTTP handlers and response types for the authentication service.
package handler

import (
	"auth/internal/dto"
	"auth/internal/service"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

var (
	// BadRequest is the error code for bad request responses.
	BadRequest = "badRequest"
	// ValidationError is the error code for validation errors.
	ValidationError = "validationError"
	// Conflict is the error code for conflict responses, e.g., duplicate email.
	Conflict = "conflict"
	// InternalError is the error code for internal server errors.
	InternalError = "internalError"
	// Unauthorized is the error code for unauthorized access.
	Unauthorized = "unauthorized"
	// NotFound is the error code for not found responses.
	NotFound = "notNound"
)

// AuthHandler handles HTTP requests for authentication and user management.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc}
}

// Register godoc
// @Summary 회원가입
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.RegisterRequest true "회원가입 정보"
// @Success 201 {object} APIResponse "예시: {\"success\":true,\"code\":201,\"message\":\"회원가입이 완료되었습니다.\",\"data\":{\"id\":1,\"email\":\"user@example.com\"}}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"필수 입력값이 누락되었습니다.\",\"data\":null}"
// @Failure 409 {object} APIResponse "예시: {\"success\":false,\"code\":409,\"message\":\"email already exists\",\"data\":null}"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(dto.RegisterRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("Register: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request body"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("Register: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	result, err := h.authService.RegisterUser(c.Context(), req)
	if err != nil {
		if err.Error() == "email already exists" {
			slog.Warn("Register: email exists", "email", req.Email)
			return c.Status(fiber.StatusConflict).JSON(NewAPIError(fiber.StatusConflict, Conflict, "email already exists"))
		}
		slog.Error("Register: internal error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(NewAPIError(fiber.StatusInternalServerError, InternalError, "internal error"))
	}
	slog.Info("User registered", "email", req.Email)
	return c.Status(fiber.StatusCreated).JSON(NewAPISuccess(result, fiber.StatusCreated, "회원가입이 완료되었습니다."))
}

// Login godoc
// @Summary 로그인
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.LoginRequest true "로그인 정보"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"로그인 성공\",\"data\":{\"accessToken\":\"...\",\"refreshToken\":\"...\"}}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"invalid credentials\",\"data\":null}"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(dto.LoginRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("User login invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("Login: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	deviceInfo := c.Get("User-Agent")
	result, err := h.authService.Login(c.Context(), req, deviceInfo)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "invalid password" {
			slog.Warn("Login failed", "email", req.Email, "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, "invalid credentials"))
		}
		slog.Error("Login: internal error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(NewAPIError(fiber.StatusInternalServerError, InternalError, "internal error"))
	}
	slog.Info("User login success", "email", req.Email)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(result, fiber.StatusOK, "로그인 성공"))
}

// RefreshToken godoc
// @Summary JWT 토큰 재발급
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.RefreshTokenRequest true "리프레시 토큰"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"토큰 재발급 성공\",\"data\":{\"accessToken\":\"...\",\"refreshToken\":\"...\"}}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid payload\",\"data\":null}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"unauthorized\",\"data\":null}"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid payload"))
	}
	accessToken, refreshToken, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, err.Error()))
	}
	resp := NewAPISuccess(fiber.Map{"accessToken": accessToken, "refreshToken": refreshToken}, fiber.StatusOK, "토큰 재발급 성공")
	return c.JSON(resp)
}

// FindEmail godoc
// @Summary 이메일(아이디) 찾기
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.FindEmailRequest true "휴대폰 번호"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"이메일 찾기 성공\",\"data\":{\"email\":\"user@example.com\"}}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Failure 404 {object} APIResponse "예시: {\"success\":false,\"code\":404,\"message\":\"not found\",\"data\":null}"
// @Router /auth/email/recover [post]
func (h *AuthHandler) FindEmail(c *fiber.Ctx) error {
	req := new(dto.FindEmailRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("FindEmail: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("FindEmail: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	result, err := h.authService.FindEmail(c.Context(), req)
	if err != nil {
		slog.Warn("FindEmail failed", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(NewAPIError(fiber.StatusNotFound, NotFound, "not found"))
	}
	slog.Info("FindEmail success", "phone", req.PhoneNumber)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(result, fiber.StatusOK, "이메일 찾기 성공"))
}

// ForgotPassword godoc
// @Summary 비밀번호 찾기(메일 발송)
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.ForgotPasswordRequest true "이메일"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"password reset email sent\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Router /auth/password/forgot [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	req := new(dto.ForgotPasswordRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("ForgotPassword: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("ForgotPassword: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	_ = h.authService.ForgotPassword(c.Context(), req.Email)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "password reset email sent"))
}

// ResetPassword godoc
// @Summary 비밀번호 재설정
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.ResetPasswordRequest true "비밀번호 재설정 정보"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"password reset successful\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid or expired token\",\"data\":null}"
// @Router /auth/password/reset [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	req := new(dto.ResetPasswordRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("ResetPassword: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("ResetPassword: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	err := h.authService.ResetPassword(c.Context(), req.Token, req.NewPassword)
	if err != nil {
		slog.Warn("ResetPassword failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid or expired token"))
	}
	slog.Info("ResetPassword success")
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "password reset successful"))
}

// Logout godoc
// @Summary 로그아웃
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body dto.LogoutRequest true "리프레시 토큰"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"logout successful\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"refresh token required\",\"data\":null}"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	req := new(dto.LogoutRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("Logout: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if req.RefreshToken == "" {
		slog.Warn("Logout: missing refresh token")
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "refresh token required"))
	}
	userID, deviceInfo, err := h.authService.JwtSvc().ValidateRefreshToken(req.RefreshToken)
	if err == nil {
		_ = h.authService.Logout(c.Context(), userID, req.RefreshToken, deviceInfo)
		slog.Info("Logout success", "userID", userID)
	} else {
		slog.Warn("Logout: invalid refresh token", "error", err)
	}
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "logout successful"))
}

// GetProfile godoc
// @Summary 내 프로필 조회
// @Tags User
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"프로필 조회 성공\",\"data\":{\"id\":1,\"email\":\"user@example.com\"}}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"unauthorized\",\"data\":null}"
// @Failure 404 {object} APIResponse "예시: {\"success\":false,\"code\":404,\"message\":\"profile not found\",\"data\":null}"
// @Router /users/me [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		slog.Warn("GetProfile: unauthorized access")
		return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, "unauthorized"))
	}
	profile, err := h.authService.GetProfile(c.Context(), userID)
	if err != nil {
		slog.Warn("GetProfile failed", "userID", userID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(NewAPIError(fiber.StatusNotFound, NotFound, "profile not found"))
	}
	slog.Info("GetProfile success", "userID", userID)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(profile, fiber.StatusOK, "프로필 조회 성공"))
}

// UpdateProfile godoc
// @Summary 내 프로필 수정
// @Tags User
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.UpdateProfileRequest true "프로필 정보"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"profile updated successfully\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"unauthorized\",\"data\":null}"
// @Router /users/me [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		slog.Warn("UpdateProfile: unauthorized access")
		return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, "unauthorized"))
	}
	req := new(dto.UpdateProfileRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("UpdateProfile: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("UpdateProfile: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	_, err := h.authService.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		slog.Warn("UpdateProfile failed", "userID", userID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, err.Error()))
	}
	slog.Info("UpdateProfile success", "userID", userID)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "profile updated successfully"))
}

// DeleteProfile godoc
// @Summary 회원 탈퇴(소프트 삭제)
// @Tags User
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.DeleteProfileRequest true "비밀번호 확인"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"account deleted (soft delete)\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"unauthorized\",\"data\":null}"
// @Router /users/me [delete]
func (h *AuthHandler) DeleteProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		slog.Warn("DeleteProfile: unauthorized access")
		return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, "unauthorized"))
	}
	req := new(dto.DeleteProfileRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("DeleteProfile: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("DeleteProfile: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	err := h.authService.CheckPassword(c.Context(), userID, req.CurrentPassword)
	if err != nil {
		slog.Warn("DeleteProfile: password check failed", "userID", userID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, err.Error()))
	}
	if err := h.authService.DeleteProfile(c.Context(), userID); err != nil {
		slog.Error("DeleteProfile failed", "userID", userID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(NewAPIError(fiber.StatusInternalServerError, InternalError, "internal error"))
	}
	slog.Info("DeleteProfile success", "userID", userID)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "account deleted (soft delete)"))
}

// ChangePassword godoc
// @Summary 내 비밀번호 변경
// @Tags User
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.ChangePasswordRequest true "비밀번호 변경 정보"
// @Success 200 {object} APIResponse "예시: {\"success\":true,\"code\":200,\"message\":\"password changed successfully\",\"data\":null}"
// @Failure 400 {object} APIResponse "예시: {\"success\":false,\"code\":400,\"message\":\"invalid request\",\"data\":null}"
// @Failure 401 {object} APIResponse "예시: {\"success\":false,\"code\":401,\"message\":\"unauthorized\",\"data\":null}"
// @Router /users/me/password [put]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		slog.Warn("ChangePassword: unauthorized access")
		return c.Status(fiber.StatusUnauthorized).JSON(NewAPIError(fiber.StatusUnauthorized, Unauthorized, "unauthorized"))
	}
	req := new(dto.ChangePasswordRequest)
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("ChangePassword: invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, "invalid request"))
	}
	if err := Validate.Struct(req); err != nil {
		slog.Warn("ChangePassword: validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, ValidationError, err.Error()))
	}
	err := h.authService.ChangePassword(c.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		slog.Warn("ChangePassword failed", "userID", userID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(NewAPIError(fiber.StatusBadRequest, BadRequest, err.Error()))
	}
	slog.Info("ChangePassword success", "userID", userID)
	return c.Status(fiber.StatusOK).JSON(NewAPISuccess(nil, fiber.StatusOK, "password changed successfully"))
}
