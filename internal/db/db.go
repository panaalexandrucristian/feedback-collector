package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Migration represents a database migration
type Migration struct {
	ID          int
	Name        string
	Description string
	FilePath    string
}

// MigrationRecord keeps track of which migrations have been executed
type MigrationRecord struct {
	ID          uint   `gorm:"primaryKey"`
	MigrationID int    `gorm:"uniqueIndex"`
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"size:255"`
	ExecutedAt  time.Time
}

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("feature_collector.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Create the migrations table if it doesn't exist
	DB.AutoMigrate(&MigrationRecord{})

	// Run migrations
	runSQLMigrations()
}

// runSQLMigrations runs SQL migrations from the migrations directory
func runSQLMigrations() {
	fmt.Println("Running SQL migrations...")

	// Get list of applied migrations
	var appliedMigrations []MigrationRecord
	if err := DB.Order("migration_id").Find(&appliedMigrations).Error; err != nil {
		panic(fmt.Errorf("failed to fetch applied migrations: %w", err))
	}

	appliedIDs := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedIDs[migration.MigrationID] = true
	}

	// Scan migration files
	migrations := scanMigrationFiles()

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	// Run pending migrations
	for _, migration := range migrations {
		if !appliedIDs[migration.ID] {
			fmt.Printf("Applying migration %d: %s\n", migration.ID, migration.Name)

			// Read SQL content from file
			sql, err := ioutil.ReadFile(migration.FilePath)
			if err != nil {
				panic(fmt.Errorf("failed to read migration file %s: %w", migration.FilePath, err))
			}

			// Begin transaction
			tx := DB.Begin()

			// Execute migration
			if err := tx.Exec(string(sql)).Error; err != nil {
				tx.Rollback()
				panic(fmt.Errorf("migration %d failed: %w", migration.ID, err))
			}

			// Record migration
			if err := tx.Create(&MigrationRecord{
				MigrationID: migration.ID,
				Name:        migration.Name,
				Description: migration.Description,
				ExecutedAt:  time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				panic(fmt.Errorf("failed to record migration %d: %w", migration.ID, err))
			}

			// Commit transaction
			tx.Commit()

			fmt.Printf("Migration %d completed successfully\n", migration.ID)
		} else {
			fmt.Printf("Migration %d already applied\n", migration.ID)
		}
	}
}

// scanMigrationFiles scans for migration files in the migrations directory
func scanMigrationFiles() []Migration {
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		panic(fmt.Errorf("failed to read migrations directory: %w", err))
	}

	// Regular expression to match migration files: 001_description.sql
	re := regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_]+)\.sql$`)

	var migrations []Migration

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(file.Name())
		if len(matches) != 3 {
			// Not a migration file, skip
			continue
		}

		id, err := strconv.Atoi(matches[1])
		if err != nil {
			fmt.Printf("Warning: Invalid migration ID in file %s\n", file.Name())
			continue
		}

		name := matches[2]
		description := strings.ReplaceAll(name, "_", " ")

		migrations = append(migrations, Migration{
			ID:          id,
			Name:        name,
			Description: description,
			FilePath:    filepath.Join(migrationsDir, file.Name()),
		})
	}

	return migrations
}

func GetDB() *gorm.DB {
	return DB
}
