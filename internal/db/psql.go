package db

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB is a singleton struct that holds the database connection and prepared statements.
type DB struct {
	conn *sql.DB

	// Prepared statements
	getPlayerStmt *sql.Stmt
}

var (
	instance *DB
	once     sync.Once
)

// GetDB returns a singleton database connection
func GetDB() *DB {
	once.Do(func() {
		// Connect to database
		connStr := os.Getenv("PSQL_CONNECTION_URL")
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}

		// Test the connection
		err = db.Ping()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}

		// Prepare the getPlayer statement
		getPlayerStmt, err := db.Prepare("SELECT email, last_signed_in_at from players WHERE id = $1")
		if err != nil {
			log.Fatal("Failed to prepare getUserStmt:", err)
		}

		// Create DB instance
		instance = &DB{
			conn:          db,
			getPlayerStmt: getPlayerStmt,
		}
	})
	return instance
}

// GetPlayer returns a player by player id
func (db *DB) GetPlayer(id int) (*casino.Player, error) {
	var email string
	var lastSignedInAt time.Time
	err := db.getPlayerStmt.QueryRow(id).Scan(&email, &lastSignedInAt)

	player := &casino.Player{
		Email:          email,
		LastSignedInAt: lastSignedInAt,
	}
	return player, err
}

// Close the database connection
func (db *DB) Close() error {
	if err := db.getPlayerStmt.Close(); err != nil {
		return err
	}
	return db.conn.Close()
}
