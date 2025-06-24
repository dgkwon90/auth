package database_test

import (
	"os"
	"path/filepath"
	"testing"

	"auth/pkg/database"

	"github.com/stretchr/testify/assert"
)

func Test_Sqlite연결실패(t *testing.T) {
	// 존재하지 않는 경로로 연결 시도
	invalidPath := "/invalid/path/to/sqlite.db"
	err := database.ConnectSqlite(invalidPath)
	assert.NotNil(t, err, "Expected an error but got nil")
}

func Test_Sqlite연결성공(t *testing.T) {
	// 임시 파일로 sqlite 연결
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite.db")
	defer func() {
		_ = os.Remove(tmpFile)
	}()
	err := database.ConnectSqlite(tmpFile)
	assert.Nil(t, err, "Expected no error but got one")
}

func Test_Sqlite연결후반환(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite2.db")
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	err := database.ConnectSqlite(tmpFile)
	assert.Nil(t, err, "Expected no error but got one")

	conn := database.GetSqliteConn()
	assert.NotNil(t, conn, "Expected a non-nil sqlite connection")
}
