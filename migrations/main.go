package main

import (
	"flag"
	"log"
	"sage-backend/internal/shared/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	up := flag.Bool("up", false, "Migrate up")
	down := flag.Bool("down", false, "Migrate down (1 step)")
	reset := flag.Bool("reset", false, "Reset migration state (force to -1)")
	forceVersion := flag.Int("force", -999, "Force a specific migration version (clears dirty state)")
	flag.Parse()

	cfg := config.SetupAPI()

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
		log.Fatal("No action specified. Use --up, --down, or --force <version>")
	}
}
