package utils_test

import (
	"auth/pkg/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HashPassword생성(t *testing.T) {
	password := "test1234"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
	assert.NotEqual(t, password, hashedPassword, "Hashed password should not be the same as the original password")
}

func Test_CheckPasswordHash성공(t *testing.T) {
	password := "test1234"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	isValid := utils.CheckPasswordHash(password, hashedPassword)
	assert.True(t, isValid, "Password should match the hashed password")
}
