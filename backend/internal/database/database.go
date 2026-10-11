package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sukamoto/wedding-invitation/internal/config"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func Open(cfg *config.Config) (*sql.DB, error) {
	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite is best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB, migrationsPath string) error {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`); err != nil {
		return fmt.Errorf("create migration tracking table: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		var applied int
		if err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, entry.Name()).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if applied != 0 {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsPath, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", entry.Name(), err)
		}
		if strings.TrimSpace(string(content)) != "" {
			if _, err := tx.Exec(string(content)); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("exec migration %s: %w", entry.Name(), err)
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
			entry.Name(), time.Now().UTC().Format(time.RFC3339)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// Seed inserts default packages, sample templates, and bootstrap admin if missing.
func Seed(db *sql.DB, cfg *config.Config) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Packages
	packages := []struct {
		slug, name string
		price      int64
		features   string
	}{
		{"silver", "Silver", 149000, `["template_tetap","foto_couple","countdown","rsvp_dasar","musik_latar","galeri_terbatas"]`},
		{"gold", "Gold", 349000, `["semua_fitur_silver","layout_opsional","custom_warna","multiple_event","live_rsvp","guestbook"]`},
		{"platinum", "Platinum VIP", 0, `["desain_custom","konsultasi","revisi_terbatas","masa_aktif_panjang"]`},
	}

	for _, p := range packages {
		var exists int
		err := db.QueryRow(`SELECT COUNT(1) FROM packages WHERE slug = ?`, p.slug).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			_, err = db.Exec(`
				INSERT INTO packages (slug, name, price, currency, features_json, active, created_at, updated_at)
				VALUES (?, ?, ?, 'IDR', ?, 1, ?, ?)`,
				p.slug, p.name, p.price, p.features, now, now)
			if err != nil {
				return fmt.Errorf("seed package %s: %w", p.slug, err)
			}
		}
	}

	// Seed every audited template in the frontend/templates directory.
	templates := []struct {
		slug, name, pkg, source, thumb, demo, meta string
	}{
		{
			"s01-floral-elegant", "S01 Classic Matcha", "silver",
			"templates/SILVER/template1_classic_matcha.html",
			"", "/demo/SILVER/template1_classic_matcha.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","quote","gallery"]}`,
		},
		{
			"s02-minimalist-white", "S02 Burgundy Minimalist", "silver",
			"templates/SILVER/template4-burgundy-minimalist.html",
			"", "/demo/SILVER/template4-burgundy-minimalist.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","quote"]}`,
		},
		{
			"s03-islamic-traditional", "S03 Islamic Traditional", "silver",
			"templates/SILVER/template8_islamic_traditional.html",
			"", "/demo/SILVER/template8_islamic_traditional.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","quote","guest_name","gallery"]}`,
		},
		{
			"g01-premium-classic", "G01 Dusty Rose", "gold",
			"templates/GOLD/dusty_rose_wedding.html",
			"", "/demo/GOLD/dusty_rose_wedding.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","quote","story","gallery"]}`,
		},
		{
			"g02-black-champagne", "G02 Black Champagne", "gold",
			"templates/GOLD/black-champagne-gold-ivory.html",
			"", "/demo/GOLD/black-champagne-gold-ivory.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","gallery"]}`,
		},
		{
			"p01-imperial-modern", "P01 Imperial Modern", "platinum",
			"templates/PLATINUM/3_Imperial-Modern.html",
			"", "/demo/PLATINUM/3_Imperial-Modern.html",
			`{"supported_fields":["bride_name","groom_name","event_date","venue","quote","guest_name","story"]}`,
		},
	}

	for _, t := range templates {
		result, err := db.Exec(`
			UPDATE templates
			SET name = ?, package_slug = ?, source_path = ?, thumbnail_path = ?,
			    demo_path = ?, metadata_json = ?, updated_at = ?
			WHERE slug = ?`,
			t.name, t.pkg, t.source, t.thumb, t.demo, t.meta, now, t.slug)
		if err != nil {
			return err
		}
		updated, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if updated == 0 {
			_, err = db.Exec(`
				INSERT INTO templates (slug, name, package_slug, source_path, thumbnail_path, demo_path, metadata_json, available, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
				t.slug, t.name, t.pkg, t.source, t.thumb, t.demo, t.meta, now, now)
			if err != nil {
				return fmt.Errorf("seed template %s: %w", t.slug, err)
			}
		}
	}

	// Bootstrap admin
	var adminCount int
	if err := db.QueryRow(`SELECT COUNT(1) FROM admins`).Scan(&adminCount); err != nil {
		return err
	}
	if adminCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}
		_, err = db.Exec(`
			INSERT INTO admins (username, password_hash, active, created_at, updated_at)
			VALUES (?, ?, 1, ?, ?)`,
			cfg.AdminUsername, string(hash), now, now)
		if err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}
	}

	return nil
}
