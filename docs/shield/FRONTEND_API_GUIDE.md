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

---

## 📈 6. Threat Severity Trends (Widget 9)

### `GET /api/v1/events/threat-trends`
Returns daily time-series threat metrics classified by security/severity levels (`critical`, `high`, `medium`, `low`, `total`) comparing the target month against a baseline/previous month period.

#### **Query Parameters**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `current_month` | `string` | current month | Target month (e.g. `2026-08`, `August`, or `8`) |
| `previous_month`| `string` | previous month | Baseline comparison month (e.g. `2026-07`, `July`, or `7`) |
| `severity`      | `string` | all | Optional filter by security level (`critical`, `high`, `medium`, `low`) |

#### **Sample Response**
```json
{
  "status": 200,
  "message": "Threat severity trends retrieved",
  "data": {
    "current_month": "August",
    "previous_month": "July",
    "severity_filter": "",
    "days": [
      {
        "day": 1,
        "critical": 3,
        "high": 8,
        "medium": 12,
        "low": 4,
        "total": 27,
        "current_month_count": 27,
        "last_month_count": 18
      },
      {
        "day": 2,
        "critical": 1,
        "high": 4,
        "medium": 7,
        "low": 2,
        "total": 14,
        "current_month_count": 14,
        "last_month_count": 20
      }
    ]
  }
}
```

---

## 🌍 7. Live Threat Origins (Geo-IP Map, Widget 10)

### `GET /api/v1/events/geo-threats` (or `GET /api/v1/geo-threats`)
Returns live Geo-IP threat origins plotted on the interactive world map, ranking top threat regions, identifying the most targeted asset (host, server, account), and summarizing total threat volume.

#### **Sample Response**
```json
{
  "status": 200,
  "message": "Geo threat origins retrieved",
  "data": {
    "total_threats": 154,
    "high_threat_region": "Russia",
    "most_targeted_asset": "finance-db-server",
    "top_targeted_assets": [
      {
        "asset": "finance-db-server",
        "count": 85,
        "type": "host"
      },
      {
        "asset": "admin-portal",
        "count": 42,
        "type": "service"
      },
      {
        "asset": "finance-vm",
        "count": 27,
        "type": "host"
      }
    ],
    "origins": [
      {
        "country": "Russia",
        "lat": 55.7558,
        "lng": 37.6173,
        "count": 85,
        "percentage": 55.19
      },
      {
        "country": "China",
        "lat": 39.9042,
        "lng": 116.4074,
        "count": 42,
        "percentage": 27.27
      },
      {
        "country": "North Korea",
        "lat": 39.0392,
        "lng": 125.7625,
        "count": 27,
        "percentage": 17.53
      }
    ]
  }
}
```
