package utils_test

import (
	"auth/pkg/utils"
	"testing"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ab@naver.com", "***@naver.com"},
		{"testuser@gmail.com", "te******@gmail.com"},
		{"a@domain.com", "***@domain.com"},
	}
	for _, tt := range tests {
		got := utils.MaskEmail(tt.input)
		if got != tt.want {
			t.Errorf("MaskEmail(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}
