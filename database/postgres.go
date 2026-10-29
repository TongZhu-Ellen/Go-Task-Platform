
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDB initializes GORM, verifies database connectivity,
// and returns both the GORM DB and underlying SQL connection pool.
func NewDB(databaseURL string) (*gorm.DB, *sql.DB, error) {
	// Initialize GORM with the PostgreSQL driver.
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	// Retrieve the underlying database/sql connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get database connection pool: %w", err)
	}

	// Verify database connectivity with a 5-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		// Clean up the connection pool if initialization fails.
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	// Return both handles so the caller can use GORM and close the pool.
	return db, sqlDB, nil
}
