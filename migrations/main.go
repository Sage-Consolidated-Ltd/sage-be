package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sage-backend/internal/shared/config"
	"sort"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func waitForDatabase(databaseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		db, err := sql.Open("postgres", databaseURL)
		if err == nil {
			err = db.Ping()
			db.Close()
			if err == nil {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for database: %w", err)
		}
		log.Println("Waiting for database to be ready...")
		time.Sleep(2 * time.Second)
	}
}

func main() {
	up := flag.Bool("up", false, "Migrate up")
	down := flag.Bool("down", false, "Migrate down (1 step)")
	reset := flag.Bool("reset", false, "Reset migration state (force to -1)")
	forceVersion := flag.Int("force", -999, "Force a specific migration version (clears dirty state)")
	seed := flag.Bool("seed", false, "Run seed files from ./seeds directory")
	seedFile := flag.String("seed-file", "", "Run a specific seed file")
	flag.Parse()

	cfg := config.SetupWorker()

	if *up || *down || *reset || *forceVersion != -999 || *seed || *seedFile != "" {
		if err := waitForDatabase(cfg.DatabaseUrl, 60*time.Second); err != nil {
			log.Fatalf("Database connection check failed: %v", err)
		}
	}

	if *up {
		log.Println("Running UP migrations...")
		m, err := migrate.New("file://migrations", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			m.Close()
			log.Fatalf("UP failed: %v", err)
		}
		m.Close()
		log.Println("Database is up to date!")
	} else if *down {
		log.Println("Running DOWN migration (reverting 1 step)...")
		m, err := migrate.New("file://migrations", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			m.Close()
			log.Fatalf("DOWN failed: %v", err)
		}
		m.Close()
		log.Println("Successfully reverted 1 migration!")
		return
	} else if *reset {
		log.Println("Resetting migration state...")
		m, err := migrate.New("file://migrations", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}
		if err := m.Force(-1); err != nil {
			m.Close()
			log.Fatalf("Reset failed: %v", err)
		}
		m.Close()
		log.Println("Reset successful! You can now run --up again.")
		return
	} else if *forceVersion != -999 {
		log.Printf("Forcing version %d...", *forceVersion)
		m, err := migrate.New("file://migrations", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}
		if err := m.Force(*forceVersion); err != nil {
			m.Close()
			log.Fatalf("Force failed: %v", err)
		}
		m.Close()
		log.Println("Force successful! You can now run --up again.")
		return
	} else if !*seed && *seedFile == "" {
		log.Fatal("No action specified. Use --up, --down, --seed, or --force <version>")
	}

	// Handle seed actions
	if *seed || *seedFile != "" {
		db, err := sqlx.Connect("postgres", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to connect to database for seeds: %v", err)
		}
		defer db.Close()

		if *seedFile != "" {
			runSeedFile(db, *seedFile)
		} else {
			runAllSeeds(db, "seeds")
		}
	}
}

func runAllSeeds(db *sqlx.DB, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read seeds directory: %v", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // ensures 001_, 002_ ordering

	if len(files) == 0 {
		log.Println("No seed files found.")
		return
	}

	for _, f := range files {
		runSeedFile(db, f)
	}
	log.Println("All seeds completed!")
}

func runSeedFile(db *sqlx.DB, path string) {
	log.Printf("Running seed: %s", path)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read seed file %s: %v", path, err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		log.Fatalf("Seed %s failed (rolled back): %v", path, err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit seed %s: %v", path, err)
	}

	fmt.Printf("  ✓ %s\n", filepath.Base(path))
}
