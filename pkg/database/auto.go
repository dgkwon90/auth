package database

import (
	"errors"
	"os"
)

// DBType represents the supported database types.
type DBType string

const (
	// Postgres is the identifier for PostgreSQL.
	Postgres DBType = "postgres"
	// Sqlite is the identifier for SQLite.
	Sqlite   DBType = "sqlite"
)

// ConnectAuto connects to the database based on dbType and dsn/path.
func ConnectAuto(dbType DBType, dsn, sqlitePath string) error {
	switch dbType {
	case Postgres:
		return Connect(dsn)
	case Sqlite:
		return ConnectSqlite(sqlitePath)
	default:
		return errors.New("unsupported db type")
	}
}

// GetDBTypeFromEnv returns the DB type from the environment variable DB_TYPE.
func GetDBTypeFromEnv() DBType {
	dbType := os.Getenv("DB_TYPE")
	if dbType == string(Sqlite) {
		return Sqlite
	}
	return Postgres
}
