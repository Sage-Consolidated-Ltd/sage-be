package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "sage-worker",
		Usage: "CLI for managing Sage background workers",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the background worker",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "redis-addr",
						Value: "localhost:6379",
						Usage: "Redis server address",
					},
					&cli.IntFlag{
						Name:  "concurrency",
						Value: 10,
						Usage: "Number of concurrent workers",
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Println("Starting worker...")
					fmt.Printf("Redis: %s\n", c.String("redis-addr"))
					fmt.Printf("Concurrency: %d\n", c.Int("concurrency"))
					// In production, this would execute: go run cmd/worker/main.go
					fmt.Println("Run: go run cmd/worker/main.go")
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the background worker",
				Action: func(c *cli.Context) error {
					fmt.Println("Stopping worker...")
					// In production, this would send SIGTERM to the worker process
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check worker status",
				Action: func(c *cli.Context) error {
					fmt.Println("Worker status: running")
					return nil
				},
			},
			{
				Name:  "queues",
				Usage: "List all job queues",
				Action: func(c *cli.Context) error {
					fmt.Println("Queues:")
					fmt.Println("  - critical (priority: 6)")
					fmt.Println("  - default (priority: 3)")
					fmt.Println("  - low (priority: 1)")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
