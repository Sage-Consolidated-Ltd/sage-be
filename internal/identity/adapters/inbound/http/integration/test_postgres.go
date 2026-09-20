package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var tables = []string{
	"industries",
	"roles",
	"permissions",
	"permission_groups",
	"permission_group_permissions",
	"users",
	"user_preferences",
	"user_notifications",
	"user_sessions",
	"organizations",
	"organization_members",
	"organization_roles",
	"organization_invites",
	"organization_settings",
	"audit_logs",
	"custom_roles",
	"security_events",
	"ingestion_jobs",
	"data_quality_scans",
	"data_sources",
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func testPostgres(t *testing.T) *db.DB {
	t.Helper()

	cfg := config.SetupTestConfig()

	root := findProjectRoot()
	migrationsDir := filepath.Join(root, "migrations")

	m, err := migrate.New("file://"+migrationsDir, cfg.DatabaseUrl)
	if err != nil {
		t.Skipf("Skipping integration test: database unavailable: %v", err)
		return nil
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Skipf("Skipping integration test: migration failed on test database: %v", err)
		return nil
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		t.Logf("migration source close warning: %v", srcErr)
	}
	if dbErr != nil {
		t.Logf("migration db close warning: %v", dbErr)
	}

	database, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to test database: %v", err)
		return nil
	}

	cleanupPostgres(t, database)
	seedPostgres(t, database)

	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})

	return database
}

func cleanupPostgres(t *testing.T, database *db.DB) {
	t.Helper()

	var dbName string
	err := database.QueryRow("SELECT current_database();").Scan(&dbName)
	if err != nil || !strings.Contains(dbName, "test") {
		t.Fatalf("CRITICAL TEST SAFETY VIOLATION: Refusing to truncate non-test database %q", dbName)
	}

	ctx := context.Background()
	_, err = database.ExecContext(ctx, fmt.Sprintf(`
		TRUNCATE TABLE %s RESTART IDENTITY CASCADE
	`, strings.Join(tables, ", ")))
	if err != nil {
		t.Fatalf("failed to clean database: %v", err)
	}
}

func seedPostgres(t *testing.T, database *db.DB) {
	t.Helper()

	root := findProjectRoot()
	seedsDir := filepath.Join(root, "seeds")
	entries, err := os.ReadDir(seedsDir)
	if err != nil {
		t.Fatalf("failed to read seeds directory: %v", err)
	}

	ctx := context.Background()
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(seedsDir, e.Name()))
			if err != nil {
				t.Fatalf("failed to read seed file %s: %v", e.Name(), err)
			}
			if _, err := database.ExecContext(ctx, string(content)); err != nil {
				t.Fatalf("failed to execute seed file %s: %v", e.Name(), err)
			}
		}
	}
}
