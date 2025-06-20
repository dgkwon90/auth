// Package service provides authentication and JWT token utilities for the authentication service.
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4" // jwt (제이더블유티, jwt)
)

// JwtService handles JWT token generation and validation.
type JwtService struct {
	accessTokenSecret  []byte
	refreshTokenSecret []byte
}

// NewJwtService creates a new JwtService.
func NewJwtService(secret string) *JwtService {
	return &JwtService{
		accessTokenSecret:  []byte(secret),
		refreshTokenSecret: []byte(secret + "-refresh"),
	}
}

// GenerateToken generates a JWT access token for the given user ID.
func (s *JwtService) GenerateToken(userID int64) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessTokenSecret)
}

// GenerateRefreshToken generates a refresh token for the given user ID and device info.
func (s *JwtService) GenerateRefreshToken(userID int64, deviceInfo string) (string, error) {
	claims := jwt.MapClaims{
		"sub": fmt.Sprint(userID),
		"dev": deviceInfo, // device info (디바이스 정보, device info)
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshTokenSecret)
}

// ValidateAccessToken validates the access token and returns the user ID.
func (s *JwtService) ValidateAccessToken(tokenString string) (userID int64, err error) {
	// RegisteredClaims 사용
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return s.accessTokenSecret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	// 만료 검증
	if claims.ExpiresAt.Before(time.Now()) {
		return 0, errors.New("access token expired")
	}
	// Subject에 저장된 userID 파싱
	var id int64
	_, err = fmt.Sscan(claims.Subject, &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ValidateRefreshToken validates the refresh token and returns the user ID and device info.
func (s *JwtService) ValidateRefreshToken(tokenString string) (userID int64, deviceInfo string, err error) {
	// MapClaims 사용
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return s.refreshTokenSecret, nil
	})
	if err != nil {
		return 0, "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", errors.New("invalid refresh token")
	}
	// 만료 검증
	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return 0, "", errors.New("refresh token expired")
	}
	// sub, dev 정보 추출
	sub, ok1 := claims["sub"].(string)
	dev, ok2 := claims["dev"].(string)
	if !ok1 || !ok2 {
		return 0, "", errors.New("invalid token claims")
	}
	var id int64
	_, err = fmt.Sscan(sub, &id)
	if err != nil {
		return 0, "", err
	}
	return id, dev, nil
}
