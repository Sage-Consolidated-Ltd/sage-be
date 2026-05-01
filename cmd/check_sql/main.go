package main

import (
	"context"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	conn, err := sql.Open("postgres", "postgres://sage_user:sage_dev_password@localhost:5433/sage_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()

	fset := token.NewFileSet()
	var queries []string
	var filePaths []string

	err = filepath.Walk("internal", func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") { return nil }

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil { return nil }

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.BasicLit:
				if x.Kind == token.STRING {
					val := strings.Trim(x.Value, "\"`")
					upper := strings.ToUpper(val)
					if strings.Contains(upper, "SELECT ") || strings.Contains(upper, "INSERT INTO ") || strings.Contains(upper, "UPDATE ") || strings.Contains(upper, "DELETE FROM ") {
						queries = append(queries, val)
						filePaths = append(filePaths, path)
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for i, q := range queries {
		if !strings.Contains(strings.ToUpper(q), "FROM") && !strings.Contains(strings.ToUpper(q), "INTO") && !strings.Contains(strings.ToUpper(q), "SET") {
			continue // skip errors matching UPDATE or strings without FROM/INTO/SET
		}
		
		_, err := conn.PrepareContext(ctx, q)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "syntax") {
				fmt.Printf("File: %s\nQuery: %s\nError: %v\n\n", filePaths[i], q, err)
			}
		}
	}
}
