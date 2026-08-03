package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sage-backend/internal/shared/config"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	up := flag.Bool("up", false, "Migrate up")
	down := flag.Bool("down", false, "Migrate down (1 step)")
	reset := flag.Bool("reset", false, "Reset migration state (force to -1)")
	forceVersion := flag.Int("force", -999, "Force a specific migration version (clears dirty state)")
	seed := flag.Bool("seed", false, "Run seed files from ./seeds directory")
	seedFile := flag.String("seed-file", "", "Run a specific seed file")
	flag.Parse()

	cfg := config.SetupWorker()

	// Handle seed actions (no migrator needed)
	if *seed || *seedFile != "" {
		db, err := sqlx.Connect("postgres", cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		if *seedFile != "" {
			runSeedFile(db, *seedFile)
		} else {
			runAllSeeds(db, "seeds")
		}
		return
	}

	// Migration actions
	m, err := migrate.New("file://migrations", cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("Warning: source close error: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("Warning: db close error: %v", dbErr)
		}
	}()

	switch {
	case *up:
		log.Println("Running UP migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("UP failed: %v", err)
		}
		log.Println("Database is up to date!")

	case *down:
		log.Println("Running DOWN migration (reverting 1 step)...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("DOWN failed: %v", err)
		}
		log.Println("Successfully reverted 1 migration!")

	case *reset:
		log.Println("Resetting migration state...")
		if err := m.Force(-1); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		log.Println("Reset successful! You can now run --up again.")

	case *forceVersion != -999:
		log.Printf("Forcing version %d...", *forceVersion)
		if err := m.Force(*forceVersion); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		log.Println("Force successful! You can now run --up again.")

	default:
		log.Fatal("No action specified. Use --up, --down, --seed, or --force <version>")
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
