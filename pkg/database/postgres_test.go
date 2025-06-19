package database_test

import (
	"auth/internal/config"
	"auth/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DB연결실패(t *testing.T) {
	// Test the database connection with invalid credentials
	dataSourceName := "user=invalid password=wrongpassword dbname=nonexistentdb sslmode=disable"
	err := database.Connect(dataSourceName)

	// Assert that an error is returned
	assert.NotNil(t, err, "Expected an error but got nil")
}

func Test_DB연결성공(t *testing.T) {
	// Test the database connection with valid credentials
	// Note: Replace with actual valid credentials for testing
	config := config.LoadConfig("E:/workspace/auth/.env")
	dataSourceName := config.DatabaseURL
	err := database.Connect(dataSourceName)

	// Assert that no error is returned
	assert.Nil(t, err, "Expected no error but got one")
}

func Test_DB연결후반환(t *testing.T) {
	// Test the database connection with valid credentials
	// Note: Replace with actual valid credentials for testing
	config := config.LoadConfig("E:/workspace/auth/.env")
	dataSourceName := config.DatabaseURL
	err := database.Connect(dataSourceName)

	// Assert that no error is returned
	assert.Nil(t, err, "Expected no error but got one")

	// Get the database connection
	dbPool := database.GetPool()
	// close the database connection after the test
	// Note: In a real-world scenario, you would want to close the connection in a deferred function or in a teardown step.
	defer dbPool.Close()

	// Assert that the database connection is not nil
	assert.NotNil(t, dbPool, "Expected a non-nil database connection")
}
