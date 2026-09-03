package models

import "sage-backend/internal/shield/domain"

type ThreatDayTrendDTO struct {
	Day               int `db:"day"`
	CurrentMonthCount int `db:"current_month_count"`
	LastMonthCount    int `db:"last_month_count"`
}

func (dto *ThreatDayTrendDTO) ToDomain() domain.ThreatDayTrend {
	return domain.ThreatDayTrend{
		Day:               dto.Day,
		CurrentMonthCount: dto.CurrentMonthCount,
		LastMonthCount:    dto.LastMonthCount,
	}
}
