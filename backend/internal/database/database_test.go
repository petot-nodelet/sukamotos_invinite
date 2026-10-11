package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateAppliesFilesOnceInOrder(t *testing.T) {
	migrationsPath := t.TempDir()
	for name, content := range map[string]string{
		"001_create.sql": "CREATE TABLE migration_test (value INTEGER NOT NULL);",
		"002_insert.sql": "INSERT INTO migration_test (value) VALUES (42);",
		"003_empty.sql":  "",
	} {
		if err := os.WriteFile(filepath.Join(migrationsPath, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "migrate.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(db, migrationsPath); err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	if err := Migrate(db, migrationsPath); err != nil {
		t.Fatalf("second migration run: %v", err)
	}

	var valueCount, migrationCount int
	if err := db.QueryRow(`SELECT COUNT(1) FROM migration_test WHERE value = 42`).Scan(&valueCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatal(err)
	}
	if valueCount != 1 {
		t.Fatalf("migration ran more than once: inserted %d rows", valueCount)
	}
	if migrationCount != 3 {
		t.Fatalf("expected all migration files, including the empty one, to be recorded; got %d", migrationCount)
	}
}

func TestProjectMigrationsAddSessionCSRFTokens(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate test source file")
	}
	migrationsPath := filepath.Join(filepath.Dir(sourceFile), "..", "..", "migrations")
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "project-migrations.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(db, migrationsPath); err != nil {
		t.Fatalf("apply project migrations: %v", err)
	}
	var csrfColumnCount int
	if err := db.QueryRow(`SELECT COUNT(1) FROM pragma_table_info('sessions') WHERE name = 'csrf_token'`).
		Scan(&csrfColumnCount); err != nil {
		t.Fatal(err)
	}
	if csrfColumnCount != 1 {
		t.Fatal("project migrations did not add sessions.csrf_token")
	}
}
