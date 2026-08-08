package main

import (
	"log"
	shieldApp "sage-backend/internal/app/shield"

	_ "sage-backend/docs/shield"
)

// @title           Sage Shield API
// @version         1.0
// @description     ## 🛡️ Sage Shield - Frontend API & Integration Guide
// @description     
// @description     ### 🔍 1. Unified AST Log Search API (`GET /api/v1/events/logs`)
// @description     Searches parsed logs across **both API-polled integrations (Okta, Entra)** and **uploaded log files (CSV, Syslog, EVTX)** using the AST Query Engine.
// @description     
// @description     **AST Query Syntax (`q` parameter)**:
// @description     - **Free-Text Phrase Search**: `q='"unauthorized access"'` or `q='"failed password"'`
// @description     - **Level Filtering**: `q='level=ERROR'` or `q='level=WARN'`
// @description     - **Source Filtering**: `q='source=123e4567-e89b-12d3-a456-426614174000'`
// @description     - **Raw Field Filtering**: `q='raw.ip_address=192.168.1.50'` or `q='raw.user_id=usr_9981'`
// @description     - **Combined Expressions**: `q='level=ERROR "unauthorized access" raw.ip_address=10.0.0.1'`
// @description     
// @description     ---
// @description     ### 📁 2. S3 Log File Upload Flow
// @description     1. Request Presigned Upload URL: `POST /api/v1/integrations/logs-data/presign-upload`
// @description     2. Upload File directly to S3 via returned presigned Form Post data.
// @description     3. Confirm & Parse File: `POST /api/v1/integrations/logs-data/confirm-upload`. Entries are automatically parsed and immediately AST-searchable!
// @description     
// @description     ---
// @description     ### 🤖 3. AI-Driven Data Quality
// @description     - `GET /api/v1/integrations/logs-data/data-quality`: Returns overall AI Quality Score & parser health.
// @description     - `GET /api/v1/integrations/logs-data/data-quality/ai-analysis`: Returns AI-detected unmapped fields & recommended parser fixes.
// @description     - `POST /api/v1/integrations/logs-data/data-quality/apply-fix`: Applies AI-suggested parser fix directly.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session_id

// @host      shield.sageconsolidated.com
// @BasePath  /api/v1
func main() {
	application, err := shieldApp.New()
	if err != nil {
		log.Fatalf("failed to initialize shield application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("shield application error: %v", err)
	}
}
