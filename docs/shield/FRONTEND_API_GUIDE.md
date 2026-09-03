# 🛡️ Sage Shield - Frontend API & Integration Guide

Welcome to the **Sage Shield API Specification**. This document outlines all endpoints, data structures, authentication methods, and query syntax for the frontend application.

---

## 🔑 1. Authentication & Base Headers

All requests require an active session or authorization header:

```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

---

## 🔍 2. Unified AST Log Search API

### `GET /api/v1/events/logs`
Searches parsed logs across **both API-polled integrations (Okta, Entra)** and **uploaded log files (CSV, Syslog, EVTX)** using the AST Query Engine.

#### **Query Parameters**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `q` | `string` | `""` | **AST Query String** (See syntax rules below) |
| `page` | `int` | `1` | Page number |
| `page_size` | `int` | `25` | Items per page (max: 200) |
| `source_id` | `uuid` | — | Filter by specific Data Source ID |
| `event_type` | `string` | — | Filter by Event Type |
| `severity` | `string` | — | Filter by Severity (`low`, `medium`, `high`, `critical`) |
| `start_time` | `string` | — | ISO-8601 Start Timestamp |
| `end_time` | `string` | — | ISO-8601 End Timestamp |

---

### 📖 **AST Search Query Syntax Guide**

The `q` parameter supports rich AST search syntax:

1. **Free-Text Phrase Search** (in double quotes):
   - `q='"unauthorized access"'`
   - `q='"failed password"'`

2. **Level Filtering**:
   - `q='level=ERROR'`
   - `q='level=WARN'`

3. **Source Filtering**:
   - `q='source=123e4567-e89b-12d3-a456-426614174000'`

4. **Raw JSON Field Filtering** (`raw.<field_name>=<value>`):
   - `q='raw.ip_address=192.168.1.50'`
   - `q='raw.user_id=usr_9981'`

5. **Combined AST Expressions**:
   - `q='level=ERROR "unauthorized access" raw.ip_address=10.0.0.1'`

---

#### **Sample Search Response**
```json
{
  "status": 200,
  "message": "Logs retrieved",
  "data": {
    "items": [
      {
        "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
        "source_id": "123e4567-e89b-12d3-a456-426614174000",
        "source": "Okta Production",
        "event_type": "user.authentication.auth_via_mfa",
        "event_category": "authentication",
        "severity": "high",
        "actor_email": "user@example.com",
        "ip_address": "192.168.1.1",
        "raw_payload": {
          "user_id": "usr_123",
          "action": "mfa_verification_failed"
        },
        "occurred_at": "2026-08-07T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 25
  }
}
```

---

## 📁 3. S3 Log File Upload Flow

To upload a raw log file (`.csv`, `.xlsx`, `.log`, EVTX) directly to S3 and trigger automated parsing & AI threat analysis:

### Step A: Request Presigned S3 Post URL
`POST /api/v1/integrations/logs-data/presign-upload`

```json
{
  "filename": "windows_security_event_log.evtx",
  "file_class": "windows_evtx"
}
```

### Step B: Upload File directly to S3
Use the returned S3 Form Post data to upload the binary file directly to S3.

### Step C: Confirm & Process File
`POST /api/v1/integrations/logs-data/confirm-upload`

```json
{
  "s3_key": "uploads/org_123/windows_security_event_log.evtx",
  "filename": "windows_security_event_log.evtx"
}
```

*File entries are automatically parsed and indexed in `parsed_logs` for immediate search!*

---

## 🤖 4. AI-Driven Data Quality Endpoints

### `GET /api/v1/integrations/logs-data/data-quality`
Returns overall organization AI Quality Score & parser health.

### `GET /api/v1/integrations/logs-data/data-quality/ai-analysis`
Returns AI-detected unmapped fields and recommended parser updates.

### `POST /api/v1/integrations/logs-data/data-quality/apply-fix`
Applies an AI-recommended fix to automatically update a parser definition.

```json
{
  "suggestion_id": "8c2deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "parser_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## 🔌 5. Provider Integrations (Okta & Entra ID)

### `POST /api/v1/integrations`
Connects an Okta or Entra ID provider for polling logs.

```json
{
  "name": "Corporate Okta",
  "provider": "okta",
  "connection_type": "polling",
  "okta": {
    "domain": "https://company.okta.com",
    "token": "secret_api_token"
  }
}
```
