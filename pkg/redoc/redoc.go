package redoc

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
)

// Config defines the configuration options for the Redocly middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// BasePath is the base URL prefix for routes (e.g. "/api/v1").
	//
	// Optional. Default: "/"
	BasePath string

	// Path is the subpath where Redocly documentation will be served (e.g. "/docs/redoc").
	//
	// Optional. Default: "/docs/redoc"
	Path string

	// FilePath is the path to the OpenAPI / Swagger JSON or YAML file on disk (e.g. "./docs/users/swagger.json").
	//
	// Required.
	FilePath string

	// SpecURL is the URL where the browser can fetch the OpenAPI / Swagger spec.
	// If empty, it defaults to path.Join(BasePath, FilePath).
	//
	// Optional.
	SpecURL string

	// Title is the title displayed in the browser tab and page header.
	//
	// Optional. Default: "Sage API Documentation"
	Title string

	// SwaggerURL is the optional URL path to the Swagger UI page (e.g. "/api/v1/docs/api-docs").
	// If set, an interactive button will be rendered allowing seamless switching to Swagger UI.
	//
	// Optional.
	SwaggerURL string

	// RedocURL is the URL to the Redoc standalone JavaScript bundle.
	//
	// Optional. Default: "https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"
	RedocURL string

	// CacheAge defines the max-age for Cache-Control header in seconds.
	//
	// Optional. Default: 3600 (1 hour)
	CacheAge int
}

// ConfigDefault defines default values for Config.
var ConfigDefault = Config{
	BasePath: "/",
	Path:     "/docs/redoc",
	Title:    "Sage API Documentation",
	RedocURL: "https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js",
	CacheAge: 3600,
}

const redocTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>{{ .Title }}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
      body {
        margin: 0;
        padding: 0;
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      }
      #redoc-container {
        height: 100vh;
      }
      .nav-switch-bar {
        position: fixed;
        top: 10px;
        right: 20px;
        z-index: 10000;
        display: flex;
        gap: 8px;
      }
      .nav-switch-btn {
        background: #0f172a;
        color: #f8fafc;
        border: 1px solid #334155;
        border-radius: 6px;
        padding: 6px 14px;
        font-size: 12px;
        font-weight: 600;
        text-decoration: none;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
        transition: all 0.2s ease;
        display: inline-flex;
        align-items: center;
        gap: 6px;
      }
      .nav-switch-btn:hover {
        background: #1e293b;
        color: #38bdf8;
        border-color: #475569;
        transform: translateY(-1px);
      }
    </style>
  </head>
  <body>
    {{ if .SwaggerURL }}
    <div class="nav-switch-bar">
      <a href="{{ .SwaggerURL }}" class="nav-switch-btn" title="Open Swagger UI">
        <span>⚡ Switch to Swagger UI</span>
      </a>
    </div>
    {{ end }}
    <div id="redoc-container"></div>
    <script src="{{ .RedocURL }}"></script>
    <script>
      Redoc.init(
        '{{ .SpecURL }}',
        {
          scrollYOffset: 0,
          hideDownloadButton: false,
          disableSearch: false,
          expandResponses: '200,201',
          requiredPropsFirst: true,
          sortTagsAlphabetically: false,
          sortOperationsAlphabetically: false,
          nativeScrollbars: true,
          theme: {
            colors: {
              primary: {
                main: '#2563eb'
              },
              success: {
                main: '#16a34a'
              },
              warning: {
                main: '#ca8a04'
              },
              error: {
                main: '#dc2626'
              },
              text: {
                primary: '#1e293b',
                secondary: '#64748b'
              },
              http: {
                get: '#2563eb',
                post: '#16a34a',
                put: '#d97706',
                patch: '#0d9488',
                delete: '#dc2626'
              }
            },
            typography: {
              fontSize: '14px',
              fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
              headings: {
                fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
                fontWeight: '600'
              },
              code: {
                fontFamily: 'JetBrains Mono, monospace',
                fontSize: '13px'
              }
            },
            sidebar: {
              width: '290px',
              backgroundColor: '#f8fafc',
              textColor: '#0f172a',
              activeTextColor: '#2563eb'
            },
            rightPanel: {
              backgroundColor: '#0f172a',
              textColor: '#ffffff'
            }
          }
        },
        document.getElementById('redoc-container')
      );
    </script>
  </body>
</html>
`

// GenerateHTML renders the standalone HTML representation of Redocly.
func GenerateHTML(cfg Config) ([]byte, error) {
	tmpl, err := template.New("redoc").Parse(redocTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redoc template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, fmt.Errorf("failed to execute redoc template: %w", err)
	}

	return buf.Bytes(), nil
}

// New creates a Fiber middleware handler that serves Redocly API documentation.
func New(config ...Config) fiber.Handler {
	cfg := ConfigDefault
	if len(config) > 0 {
		cfg = config[0]
		if cfg.BasePath == "" {
			cfg.BasePath = ConfigDefault.BasePath
		}
		if cfg.Path == "" {
			cfg.Path = ConfigDefault.Path
		}
		if cfg.Title == "" {
			cfg.Title = ConfigDefault.Title
		}
		if cfg.RedocURL == "" {
			cfg.RedocURL = ConfigDefault.RedocURL
		}
		if cfg.CacheAge == 0 {
			cfg.CacheAge = ConfigDefault.CacheAge
		}
	}

	// Determine SpecURL
	if cfg.SpecURL == "" {
		if cfg.FilePath != "" {
			cfg.SpecURL = path.Join(cfg.BasePath, cfg.FilePath)
		} else {
			cfg.SpecURL = path.Join(cfg.BasePath, "swagger.json")
		}
	}

	// Read spec into memory if file exists
	var rawSpec []byte
	if cfg.FilePath != "" {
		content, err := os.ReadFile(cfg.FilePath)
		if err == nil {
			rawSpec = content
		}
	}

	htmlContent, err := GenerateHTML(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to generate redoc html: %v", err))
	}

	fullRedocPath := path.Join(cfg.BasePath, cfg.Path)

	return func(c *fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		currentPath := c.Path()

		// Serve Redocly UI
		if currentPath == fullRedocPath || currentPath == "/docs" || currentPath == "/redoc" {
			c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
			c.Set(fiber.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", cfg.CacheAge))
			return c.Send(htmlContent)
		}

		// Serve spec file if matched and available
		if len(rawSpec) > 0 && (currentPath == cfg.SpecURL || currentPath == path.Join(cfg.BasePath, cfg.FilePath)) {
			if strings.HasSuffix(currentPath, ".yaml") || strings.HasSuffix(currentPath, ".yml") {
				c.Set(fiber.HeaderContentType, "application/yaml")
			} else {
				c.Set(fiber.HeaderContentType, "application/json")
			}
			c.Set(fiber.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", cfg.CacheAge))
			return c.Send(rawSpec)
		}

		return c.Next()
	}
}
