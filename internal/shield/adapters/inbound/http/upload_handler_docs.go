package http

import (
	_ "sage-backend/internal/shared/response"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
)

type UploadPresignResponse = dto.PresignUploadResponse
type ConfirmUploadResponse = domain.LogFile

// @Summary		Create Upload Presign
// @Description	Initiates a direct-to-S3 upload flow by generating a presigned POST URL.
// @Description
// @Description	## Flow
// @Description	1. Call this endpoint with the file metadata (name, size, content type)
// @Description	2. Use the returned `post.url` and `post.fields` to upload the file directly to S3 as `multipart/form-data`
// @Description	3. The `file` field must be the **last** field in the multipart body
// @Description	4. On S3 success (HTTP 204), call `/upload/complete` with the `key` and `ETag` from the S3 response headers
// @Description
// @Description	## Rules
// @Description	- Supported extensions: `.json`, `.csv`, `.xml`, `.gz`, `.zip`, `.tar`, `.pcap`, `.log`, `.evt`, `.evtx`, `.txt`, `.yml`, `.yaml`
// @Description	- Maximum file size is enforced via the `content-length-range` condition in the policy
// @Description	- The presign URL expires in 24 hours — complete the upload before expiry
// @Description	- The returned `key` must be passed exactly as-is to the complete endpoint
// @Description	- Do not modify or add extra fields to the S3 POST — the policy will reject the request with HTTP 403
// @Tags		Logs & Data
// @Accept		json
// @Produce		json
// @Security	SessionAuth
// @Param		request	body		dto.UploadLogRequest	true	"File metadata"
// @Success		200		{object}	UploadPresignResponse	"Presign response containing S3 POST URL, fields, key and expiry"
// @Failure		400		{object}	response.Response		"Invalid request body or unsupported file type"
// @Failure		401		{object}	response.Response		"Unauthorized"
// @Failure		413		{object}	response.Response		"File size exceeds maximum allowed"
// @Failure		500		{object}	response.Response		"Internal server error"
// @Router		/integrations/logs-data/upload/presign [post]
func _UploadLog() {}

// @Summary		Complete Upload
// @Description	Confirms a direct S3 upload by verifying the file exists in S3 and storing the log file record.
// @Description
// @Description	## Flow
// @Description	Call this endpoint after receiving HTTP 204 from the S3 POST upload.
// @Description
// @Description	## Rules
// @Description	- `key` must match exactly what was returned by `/upload/presign`
// @Description	- `etag` must match the `ETag` response header from the S3 POST — include surrounding quotes e.g. `"abc123"`
// @Description	- The file must exist in S3 at the given key or the request will fail with 404
// @Description	- `metadata.source_type` is required — describes the log source e.g. `firewall`, `nginx`, `windows_security`
// @Description	- The upload can only be confirmed once — subsequent calls with the same key will fail
// @Description	- `source_id` is optional — links the log file to an existing data source in the system
// @Tags		Logs & Data
// @Accept		json
// @Produce		json
// @Security	SessionAuth
// @Param		request	body		dto.UploadCompleteRequest	true	"Upload complete request"
// @Success		200		{object}	ConfirmUploadResponse		"Confirmed log file record"
// @Failure		400		{object}	response.Response			"Invalid request body or ETag mismatch"
// @Failure		401		{object}	response.Response			"Unauthorized"
// @Failure		404		{object}	response.Response			"File not found in S3"
// @Failure		409		{object}	response.Response			"Upload already confirmed"
// @Failure		500		{object}	response.Response			"Internal server error"
// @Router		/integrations/logs-data/upload/complete [post]
func _UploadComplete() {}
