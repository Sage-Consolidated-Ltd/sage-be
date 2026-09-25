package redoc

import (
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateHTML(t *testing.T) {
	cfg := Config{
		Title:      "Test API Docs",
		SpecURL:    "/api/v1/swagger.json",
		SwaggerURL: "/api/v1/docs/swagger",
		RedocURL:   "https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js",
	}

	html, err := GenerateHTML(cfg)
	require.NoError(t, err)
	content := string(html)

	assert.Contains(t, content, "<title>Test API Docs</title>")
	assert.Contains(t, content, "Redoc.init(")
	assert.Contains(t, content, "'/api/v1/swagger.json'")
	assert.Contains(t, content, `href="/api/v1/docs/swagger"`)
	assert.Contains(t, content, "Switch to Swagger UI")
	assert.Contains(t, content, "sortTagsAlphabetically: false")
}

func TestRedocMiddleware(t *testing.T) {
	// Create a temporary swagger.json file for test
	tempFile, err := os.CreateTemp("", "swagger-*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	sampleSpec := `{"swagger":"2.0","info":{"title":"Test API"}}`
	_, err = tempFile.WriteString(sampleSpec)
	require.NoError(t, err)
	tempFile.Close()

	app := fiber.New()

	redocHandler := New(Config{
		BasePath:   "/api/v1",
		Path:       "/docs/redoc",
		FilePath:   tempFile.Name(),
		SpecURL:    "/api/v1/docs/users/swagger.json",
		Title:      "Sage Test Documentation",
		SwaggerURL: "/api/v1/docs/api-docs",
	})
	app.Use(redocHandler)

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	// 1. Test accessing configured redoc UI path (/api/v1/docs/redoc)
	req := httptest.NewRequest("GET", "/api/v1/docs/redoc", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "<title>Sage Test Documentation</title>")
	assert.Contains(t, string(body), "/api/v1/docs/users/swagger.json")

	// 2. Test accessing /docs alias
	reqDocs := httptest.NewRequest("GET", "/docs", nil)
	respDocs, err := app.Test(reqDocs)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respDocs.StatusCode)

	// 3. Test accessing /redoc alias
	reqRedoc := httptest.NewRequest("GET", "/redoc", nil)
	respRedoc, err := app.Test(reqRedoc)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respRedoc.StatusCode)

	// 4. Test accessing the spec URL
	reqSpec := httptest.NewRequest("GET", "/api/v1/docs/users/swagger.json", nil)
	respSpec, err := app.Test(reqSpec)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respSpec.StatusCode)
	specBody, _ := io.ReadAll(respSpec.Body)
	assert.True(t, strings.Contains(string(specBody), "Test API"))

	// 5. Test pass-through to API endpoints
	reqHealth := httptest.NewRequest("GET", "/api/v1/health", nil)
	respHealth, err := app.Test(reqHealth)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respHealth.StatusCode)
	healthBody, _ := io.ReadAll(respHealth.Body)
	assert.Equal(t, "healthy", string(healthBody))
}
