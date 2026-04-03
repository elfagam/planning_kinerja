package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load()

	// Robustness: Tunggu sampai database siap (penting untuk Railway/Docker)
	log.Printf("[MIGRATE] Memeriksa koneksi ke host %s...", cfg.DBHost)
	if err := database.PingMySQL(cfg); err != nil {
		log.Fatalf("[MIGRATE] GAGAL: tidak bisa terhubung ke database setelah beberapa kali percobaan: %v", err)
	}
	log.Printf("[MIGRATE] Database terhubung, mulai proses migrasi")

	dsn := withMultiStatements(cfg.MySQLDSN)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		err = database.WrapConnectionError("sql open", fmt.Errorf("failed to open mysql connection: %w", err))
		if database.IsConnectionError(err) {
			log.Fatalf("%v; %s", err, database.ConnectionFailureHint())
		}
		log.Fatalf("failed to open mysql connection: %v", err)
	}
	defer db.Close()

	migrationDir := os.Getenv("MIGRATIONS_DIR")
	if migrationDir == "" {
		migrationDir = "migrations"
	}

	if err := ensureMigrationTable(db); err != nil {
		log.Fatalf("failed to ensure migration table: %v", err)
	}

	paths, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		log.Fatalf("failed to read migration files: %v", err)
	}
	sort.Strings(paths)

	if len(paths) == 0 {
		log.Printf("no migration files found in %s", migrationDir)
		return
	}

	for _, path := range paths {
		name := filepath.Base(path)
		applied, err := isApplied(db, name)
		if err != nil {
			log.Fatalf("failed to check migration %s: %v", name, err)
		}
		if applied {
			log.Printf("skip %s (already applied)", name)
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read %s: %v", name, err)
		}

		if err := applyMigration(db, name, string(content)); err != nil {
			log.Fatalf("failed to apply %s: %v", name, err)
		}

		log.Printf("applied %s", name)
	}

	log.Println("all migrations completed")
}

func ensureMigrationTable(db *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_schema_migrations_name (name)
);`

	_, err := db.Exec(query)
	return err
}

func isApplied(db *sql.DB, name string) (bool, error) {
	const query = `SELECT COUNT(1) FROM schema_migrations WHERE name = ?`
	var count int
	if err := db.QueryRow(query, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyMigration(db *sql.DB, name, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(sqlContent); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func withMultiStatements(dsn string) string {
	if strings.Contains(strings.ToLower(dsn), "multistatements=") {
		return dsn
	}

	if strings.Contains(dsn, "?") {
		return fmt.Sprintf("%s&multiStatements=true", dsn)
	}

	return fmt.Sprintf("%s?multiStatements=true", dsn)
}
