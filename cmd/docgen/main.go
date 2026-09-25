package main

import (
	"log"
	"os"
	"path/filepath"

	"sage-backend/pkg/redoc"
)

type docTarget struct {
	outputDir  string
	title      string
	specURL    string
	swaggerURL string
}

func main() {
	targets := []docTarget{
		{
			outputDir:  "./docs/users",
			title:      "Sage API (Identity & Organization) Documentation",
			specURL:    "./swagger.json",
			swaggerURL: "/api/v1/docs/api-docs",
		},
		{
			outputDir:  "./docs/shield",
			title:      "Sage Shield API Documentation",
			specURL:    "./swagger.json",
			swaggerURL: "/api/v1/docs/shield-docs",
		},
		{
			outputDir:  "./docs/admin",
			title:      "Sage Admin API Documentation",
			specURL:    "./swagger.json",
			swaggerURL: "/api/v1/docs/admin-docs",
		},
	}

	for _, t := range targets {
		html, err := redoc.GenerateHTML(redoc.Config{
			Title:      t.title,
			SpecURL:    t.specURL,
			SwaggerURL: t.swaggerURL,
			RedocURL:   "https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js",
		})
		if err != nil {
			log.Fatalf("failed to generate redoc html for %s: %v", t.outputDir, err)
		}

		targetFile := filepath.Join(t.outputDir, "redoc.html")
		if err := os.WriteFile(targetFile, html, 0644); err != nil {
			log.Fatalf("failed to write %s: %v", targetFile, err)
		}
		log.Printf("Generated %s successfully", targetFile)
	}
}
