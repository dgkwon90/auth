package database

import (
	"log"
	"zombiezen.com/go/sqlite"
)

var sqliteConn *sqlite.Conn

// ConnectSqlite initializes the sqlite connection using zombiezen.com/go/sqlite
func ConnectSqlite(sqlitePath string) error {
	conn, err := sqlite.OpenConn(sqlitePath, 0)
	if err != nil {
		if conn != nil {
			if cerr := conn.Close(); cerr != nil {
				log.Printf("sqlite close error: %v", cerr)
			}
		}
		log.Printf("sqlite open error: %v", err)
		return err
	}
	// ping test
	stmt, err := conn.Prepare("SELECT 1;")
	if err != nil {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("sqlite close error: %v", cerr)
		}
		log.Printf("sqlite ping error: %v", err)
		return err
	}
	_, err = stmt.Step()
	if ferr := stmt.Finalize(); ferr != nil {
		log.Printf("sqlite finalize error: %v", ferr)
	}
	if err != nil {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("sqlite close error: %v", cerr)
		}
		log.Printf("sqlite ping error: %v", err)
		return err
	}
	sqliteConn = conn
	log.Printf("sqlite connection established (zombiezen)")
	return nil
}

// GetSqliteConn returns the current sqlite connection.
func GetSqliteConn() *sqlite.Conn {
	return sqliteConn
}
