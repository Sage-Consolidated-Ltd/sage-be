package models

import (
	"database/sql"
	"time"
)

type UserPreferences struct {
	ID                   string         `json:"id" db:"id"`
	UserID               string         `json:"user_id" db:"user_id"`
	OrganizationID       sql.NullString `json:"organization_id" db:"organization_id"`
	Theme                string         `json:"theme" db:"theme"`
	Timezone             string         `json:"timezone" db:"timezone"`
	Language             string         `json:"language" db:"language"`
	DashboardDefaultView string         `json:"dashboard_default_view" db:"dashboard_default_view"`
	TablePageSize        int            `json:"table_page_size" db:"table_page_size"`
	AutoRefreshInterval  int            `json:"auto_refresh_interval" db:"auto_refresh_interval"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

type UserPreferencesResponse struct {
	Theme                string `json:"theme"`
	Timezone             string `json:"timezone"`
	Language             string `json:"language"`
	DashboardDefaultView string `json:"dashboard_default_view"`
	TablePageSize        int    `json:"table_page_size"`
	AutoRefreshInterval  int    `json:"auto_refresh_interval"`
}

func (p *UserPreferences) ToResponse() *UserPreferencesResponse {
	return &UserPreferencesResponse{
		Theme:                p.Theme,
		Timezone:             p.Timezone,
		Language:             p.Language,
		DashboardDefaultView: p.DashboardDefaultView,
		TablePageSize:        p.TablePageSize,
		AutoRefreshInterval:  p.AutoRefreshInterval,
	}
}
