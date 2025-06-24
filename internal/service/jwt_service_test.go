package service_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"auth/internal/service"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func Test_JwtService_AccessToken(t *testing.T) {
	jwtSvc := service.NewJwtService("test-secret")
	userID := int64(12345)

	// 토큰 생성
	token, err := jwtSvc.GenerateToken(userID)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	// 토큰 검증
	parsedID, err := jwtSvc.ValidateAccessToken(token)
	assert.Nil(t, err)
	assert.Equal(t, userID, parsedID)
}

func Test_JwtService_RefreshToken(t *testing.T) {
	jwtSvc := service.NewJwtService("test-secret")
	userID := int64(67890)
	deviceInfo := "test-device"

	// 리프레시 토큰 생성
	token, err := jwtSvc.GenerateRefreshToken(userID, deviceInfo)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	// 리프레시 토큰 검증
	parsedID, parsedDev, err := jwtSvc.ValidateRefreshToken(token)
	assert.Nil(t, err)
	assert.Equal(t, userID, parsedID)
	assert.Equal(t, deviceInfo, parsedDev)
}

func Test_JwtService_RefreshToken_Invalid(t *testing.T) {
	jwtSvc := service.NewJwtService("test-secret")

	// 잘못된 토큰
	_, _, err := jwtSvc.ValidateRefreshToken("invalid.token.value")
	assert.NotNil(t, err)
}

// 테스트용: 만료 시간 지정해서 access token 생성
func generateAccessTokenWithExpiry(secret []byte, userID int64, expiresAt time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(userID),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// 테스트용: 만료 시간 지정해서 refresh token 생성
func generateRefreshTokenWithExpiry(secret []byte, userID int64, deviceInfo string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": fmt.Sprint(userID),
		"dev": deviceInfo,
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func Test_JwtService_AccessToken_Expired(t *testing.T) {
	jwtSvc := service.NewJwtService("test-secret")
	userID := int64(12345)

	// 만료된 토큰 생성 (1초 전 만료)
	expiredAt := time.Now().Add(-1 * time.Second)
	token, err := generateAccessTokenWithExpiry(jwtSvcAccessSecret(jwtSvc), userID, expiredAt)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	// 만료된 토큰 검증
	_, err = jwtSvc.ValidateAccessToken(token)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func Test_JwtService_RefreshToken_Expired(t *testing.T) {
	jwtSvc := service.NewJwtService("test-secret")
	userID := int64(67890)
	deviceInfo := "test-device"

	// 만료된 리프레시 토큰 생성 (1초 전 만료)
	expiredAt := time.Now().Add(-1 * time.Second)
	token, err := generateRefreshTokenWithExpiry(jwtSvcRefreshSecret(jwtSvc), userID, deviceInfo, expiredAt)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	// 만료된 토큰 검증
	_, _, err = jwtSvc.ValidateRefreshToken(token)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// JwtService의 시크릿을 테스트에서만 안전하게 가져오는 헬퍼
func jwtSvcAccessSecret(s *service.JwtService) []byte {
	typeAccessor := fmt.Sprintf("%T", s)
	if typeAccessor == "*service.JwtService" {
		return getUnexportedField(s, "accessTokenSecret").([]byte)
	}
	return nil
}
func jwtSvcRefreshSecret(s *service.JwtService) []byte {
	typeAccessor := fmt.Sprintf("%T", s)
	if typeAccessor == "*service.JwtService" {
		return getUnexportedField(s, "refreshTokenSecret").([]byte)
	}
	return nil
}

// 리플렉션으로 비공개 필드 접근 (테스트에서만 사용)
func getUnexportedField(obj interface{}, field string) interface{} {
	v := reflect.ValueOf(obj).Elem().FieldByName(field)
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface()
}
