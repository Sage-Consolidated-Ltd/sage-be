package models

import "sage-backend/internal/shield/domain"

type ThreatDayTrendDTO struct {
	Day               int `db:"day"`
	Critical          int `db:"critical"`
	High              int `db:"high"`
	Medium            int `db:"medium"`
	Low               int `db:"low"`
	Total             int `db:"total"`
	CurrentMonthCount int `db:"current_month_count"`
	LastMonthCount    int `db:"last_month_count"`
}

func (dto *ThreatDayTrendDTO) ToDomain() domain.ThreatDayTrend {
	return domain.ThreatDayTrend{
		Day:               dto.Day,
		Critical:          dto.Critical,
		High:              dto.High,
		Medium:            dto.Medium,
		Low:               dto.Low,
		Total:             dto.Total,
		CurrentMonthCount: dto.CurrentMonthCount,
		LastMonthCount:    dto.LastMonthCount,
	}
}

type GeoOriginDTO struct {
	Country string `db:"country"`
	Count   int64  `db:"count"`
}

type TargetedAssetDTO struct {
	Asset     string `db:"asset"`
	AssetType string `db:"asset_type"`
	Count     int64  `db:"count"`
}
