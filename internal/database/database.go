package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"meeting-slot-service/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/go-sql-driver/mysql"
)

// DBCredentials represents the structure of credentials stored in Secrets Manager
type DBCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

// Database wraps the SQL database connection with lazy initialization
type Database struct {
	db            *sql.DB
	cfg           *config.Config
	once          sync.Once
	initErr       error
	migrationOnce sync.Once
	migrationErr  error
}

// New creates a new Database instance (does not connect yet)
func New(cfg *config.Config) *Database {
	return &Database{
		cfg: cfg,
	}
}

// connect performs the actual database connection (called lazily)
func (d *Database) connect() error {
	d.once.Do(func() {
		var dsn string

		// If SecretARN is provided, fetch credentials from AWS Secrets Manager
		if d.cfg.Database.SecretARN != "" {
			creds, err := getCredentialsFromSecretsManager(d.cfg.AWS.Region, d.cfg.Database.SecretARN)
			if err != nil {
				d.initErr = fmt.Errorf("failed to get credentials from secrets manager: %w", err)
				return
			}
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
				creds.Username, creds.Password, creds.Host, creds.Port, creds.DBName)
		} else {
			dsn = d.cfg.Database.DSN()
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			d.initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// Set connection pool settings
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			d.initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		log.Println("Successfully connected to database (lazy initialization)")
		d.db = db
	})

	return d.initErr
}

// DB returns the underlying sql.DB connection, connecting lazily if needed
func (d *Database) DB() (*sql.DB, error) {
	if err := d.connect(); err != nil {
		return nil, err
	}
	return d.db, nil
}

// getCredentialsFromSecretsManager retrieves database credentials from AWS Secrets Manager
func getCredentialsFromSecretsManager(region, secretARN string) (*DBCredentials, error) {
	ctx := context.Background()

	// Load AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(awsCfg)

	// Get secret value
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	// Parse credentials from JSON
	var creds DBCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// RunMigrations creates the database tables (runs lazily, only once)
func (d *Database) RunMigrations() error {
	d.migrationOnce.Do(func() {
		db, err := d.DB()
		if err != nil {
			d.migrationErr = err
			return
		}

		log.Println("Running database migrations...")

		migrations := []string{
			`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
			`CREATE TABLE IF NOT EXISTS events (
			id VARCHAR(50) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			organizer_id VARCHAR(50) NOT NULL,
			duration_minutes INT NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_events_organizer (organizer_id),
			INDEX idx_events_status (status),
			INDEX idx_events_deleted (deleted_at)
		)`,
			`CREATE TABLE IF NOT EXISTS proposed_slots (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_id VARCHAR(50) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			timezone VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_proposed_slots_event (event_id),
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		)`,
			`CREATE TABLE IF NOT EXISTS event_participants (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_id VARCHAR(50) NOT NULL,
			user_id VARCHAR(50) NOT NULL,
			status VARCHAR(20) DEFAULT 'invited',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE INDEX idx_event_user (event_id, user_id),
			INDEX idx_participants_event (event_id),
			INDEX idx_participants_user (user_id),
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
			`CREATE TABLE IF NOT EXISTS availability_slots (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_id VARCHAR(50) NOT NULL,
			user_id VARCHAR(50) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			timezone VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_availability_event_user (event_id, user_id),
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		}

		for _, migration := range migrations {
			if _, err := db.Exec(migration); err != nil {
				d.migrationErr = fmt.Errorf("failed to run migration: %w", err)
				return
			}
		}

		log.Println("Database migrations completed successfully")
	})

	return d.migrationErr
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
