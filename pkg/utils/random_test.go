package utils_test

import (
	"testing"

	"auth/pkg/utils"
)

func TestGenerateRandomString_Length(t *testing.T) {
	n := 16
	s := utils.GenerateRandomString(n)
	if len(s) != n {
		t.Errorf("expected length %d, got %d", n, len(s))
	}
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	n := 16
	s1 := utils.GenerateRandomString(n)
	s2 := utils.GenerateRandomString(n)
	if s1 == s2 {
		t.Errorf("expected different strings, got same: %s", s1)
	}
}

func TestGenerateRandomString_ZeroLength(t *testing.T) {
	s := utils.GenerateRandomString(0)
	if s != "" {
		t.Errorf("expected empty string, got: %s", s)
	}
}
