package main

import (
	"log"

	"sage-backend/internal/worker"
)

func main() {
	w, err := worker.New()
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}

	if err := w.Run(); err != nil {
		log.Fatalf("worker execution error: %v", err)
	}
}
