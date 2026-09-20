package postgres

import (
	"testing"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToDomainAlert_Mapping(t *testing.T) {
	id := uuid.New()
	orgID := uuid.New()
	host := "HOST-01"
	account := "admin"
	ip := "192.168.1.5"
	now := time.Now()

	row := &alertDB{
		ID:             id,
		OrganizationID: orgID,
		ThreatLabel:    "Brute_Force_Failed_Login",
		MITRE:          "T1110.001",
		LogSource:      "Security",
		EventID:        "4625",
		EntityHost:     &host,
		EntityAccount:  &account,
		EntityIP:       &ip,
		Context:        db.JSONMap{"logon_type": 3},
		DetectedAt:     now,
		CreatedAt:      now,
	}

	domainAlert := toDomainAlert(row)

	assert.Equal(t, id, domainAlert.ID)
	assert.Equal(t, orgID, domainAlert.OrganizationID)
	assert.Equal(t, "Brute_Force_Failed_Login", domainAlert.ThreatLabel)
	assert.Equal(t, "T1110.001", domainAlert.MITRE)
	assert.Equal(t, "Security", domainAlert.LogSource)
	assert.Equal(t, "4625", domainAlert.EventID)
	assert.Equal(t, host, domainAlert.EntityHost)
	assert.Equal(t, account, domainAlert.EntityAccount)
	assert.Equal(t, ip, domainAlert.EntityIP)
	assert.Equal(t, 3, domainAlert.Context["logon_type"])
}

func TestStringToPtr(t *testing.T) {
	assert.Nil(t, stringToPtr(""))
	ptr := stringToPtr("hello")
	assert.NotNil(t, ptr)
	assert.Equal(t, "hello", *ptr)
}

func TestAlertRepository_SaveAlert_NilHandling(t *testing.T) {
	repo := NewAlertRepository(nil)
	err := repo.SaveAlert(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot save nil alert")

	// Empty bulk save returns nil
	err = repo.BulkSaveAlerts(nil, []*domain.Alert{})
	assert.NoError(t, err)
}

