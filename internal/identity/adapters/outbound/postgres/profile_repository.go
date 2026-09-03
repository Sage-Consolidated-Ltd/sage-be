package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/outbound"
)

type ProfileRepository struct {
	db.Repository
}

func NewProfileRepository(database *db.DB) outbound.ProfileRepository {
	return &ProfileRepository{
		Repository: db.NewRepository(database),
	}
}

// SQL Queries
const (
	getUserPreferencesSQL = `
		SELECT id, user_id, organization_id, theme, timezone, language, 
		       COALESCE(date_format, 'DD/MM/YYYY') as date_format,
		       dashboard_default_view, table_page_size, auto_refresh_interval, 
		       created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1 AND (organization_id = $2 OR organization_id IS NULL)
		ORDER BY organization_id DESC NULLS LAST
		LIMIT 1
	`

	upsertUserPreferencesSQL = `
		INSERT INTO user_preferences (
			user_id, organization_id, theme, timezone, language, date_format,
			dashboard_default_view, table_page_size, auto_refresh_interval
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, organization_id) 
		DO UPDATE SET
			theme = EXCLUDED.theme,
			timezone = EXCLUDED.timezone,
			language = EXCLUDED.language,
			date_format = EXCLUDED.date_format,
			dashboard_default_view = EXCLUDED.dashboard_default_view,
			table_page_size = EXCLUDED.table_page_size,
			auto_refresh_interval = EXCLUDED.auto_refresh_interval,
			updated_at = NOW()
		RETURNING id
	`

	getUserNotificationsSQL = `
		SELECT id, user_id, organization_id, email_enabled, push_enabled, slack_enabled,
		       COALESCE(product_updates, true) as product_updates,
		       COALESCE(weekly_summary, false) as weekly_summary,
		       alert_severity_threshold, notify_on_new_alert, notify_on_incident_update,
		       notify_on_playbook_execution, created_at, updated_at
		FROM user_notifications
		WHERE user_id = $1 AND (organization_id = $2 OR organization_id IS NULL)
		ORDER BY organization_id DESC NULLS LAST
		LIMIT 1
	`

	upsertUserNotificationsSQL = `
		INSERT INTO user_notifications (
			user_id, organization_id, email_enabled, push_enabled, slack_enabled,
			product_updates, weekly_summary,
			alert_severity_threshold, notify_on_new_alert, notify_on_incident_update,
			notify_on_playbook_execution
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id, organization_id)
		DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			push_enabled = EXCLUDED.push_enabled,
			slack_enabled = EXCLUDED.slack_enabled,
			product_updates = EXCLUDED.product_updates,
			weekly_summary = EXCLUDED.weekly_summary,
			alert_severity_threshold = EXCLUDED.alert_severity_threshold,
			notify_on_new_alert = EXCLUDED.notify_on_new_alert,
			notify_on_incident_update = EXCLUDED.notify_on_incident_update,
			notify_on_playbook_execution = EXCLUDED.notify_on_playbook_execution,
			updated_at = NOW()
		RETURNING id
	`

	getUserSessionsSQL = `
		SELECT id, user_id, organization_id, ip_address, user_agent, location,
		       is_revoked, created_at, last_active_at, expires_at
		FROM user_sessions
		WHERE user_id = $1 
		  AND (organization_id = $2 OR organization_id IS NULL)
		  AND is_revoked = false
		  AND expires_at > NOW()
		ORDER BY last_active_at DESC
	`

	getUserSessionByIDSQL = `
		SELECT id, user_id, organization_id, ip_address, user_agent, location,
		       is_revoked, created_at, last_active_at, expires_at
		FROM user_sessions
		WHERE id = $1 AND user_id = $2
	`

	createUserSessionSQL = `
		INSERT INTO user_sessions (
			user_id, organization_id, session_token_hash, ip_address, user_agent, 
			location, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	revokeUserSessionSQL = `
		UPDATE user_sessions 
		SET is_revoked = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	updateSessionActivitySQL = `
		UPDATE user_sessions 
		SET last_active_at = NOW()
		WHERE id = $1
	`

	getUserActivitySQL = `
		SELECT id, user_id, organization_id, action_type, resource_type, 
		       resource_id, metadata, created_at
		FROM audit_logs
		WHERE user_id = $1 AND organization_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	countUserActivitySQL = `
		SELECT COUNT(*) FROM audit_logs
		WHERE user_id = $1 AND organization_id = $2
	`

	createAuditLogSQL = `
		INSERT INTO audit_logs (
			user_id, organization_id, action_type, resource_type, 
			resource_id, metadata, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	updateLastLoginSQL = `
		UPDATE users SET last_login_at = NOW() WHERE id = $1
	`

	updateUserAvatarSQL = `
		UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2
	`
)

// GetUserPreferences retrieves user preferences for the given user and org
func (r *ProfileRepository) GetUserPreferences(ctx context.Context, userID, orgID string) (*domain.UserPreferences, error) {
	var prefs domain.UserPreferences
	err := r.Executor(ctx).GetContext(ctx, &prefs, getUserPreferencesSQL, userID, sql.NullString{String: orgID, Valid: orgID != ""})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default preferences if none exist
			return &domain.UserPreferences{
				UserID:               userID,
				OrganizationID:       sql.NullString{String: orgID, Valid: orgID != ""},
				Theme:                "system",
				Timezone:             "UTC",
				Language:             "en",
				DashboardDefaultView: "overview",
				TablePageSize:        25,
				AutoRefreshInterval:  30,
			}, nil
		}
		return nil, err
	}
	return &prefs, nil
}

// UpsertUserPreferences creates or updates user preferences
func (r *ProfileRepository) UpsertUserPreferences(ctx context.Context, userID, orgID string, prefs *domain.UserPreferences) error {
	var id string
	err := r.Executor(ctx).QueryRowxContext(ctx, upsertUserPreferencesSQL,
		userID,
		sql.NullString{String: orgID, Valid: orgID != ""},
		prefs.Theme,
		prefs.Timezone,
		prefs.Language,
		prefs.DateFormat,
		prefs.DashboardDefaultView,
		prefs.TablePageSize,
		prefs.AutoRefreshInterval,
	).Scan(&id)
	return err
}

// GetUserNotifications retrieves user notification settings
func (r *ProfileRepository) GetUserNotifications(ctx context.Context, userID, orgID string) (*domain.UserNotifications, error) {
	var notifs domain.UserNotifications
	err := r.Executor(ctx).GetContext(ctx, &notifs, getUserNotificationsSQL, userID, sql.NullString{String: orgID, Valid: orgID != ""})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default notifications if none exist
			return &domain.UserNotifications{
				UserID:                    userID,
				OrganizationID:            sql.NullString{String: orgID, Valid: orgID != ""},
				EmailEnabled:              true,
				PushEnabled:               false,
				SlackEnabled:              false,
				ProductUpdates:            true,
				WeeklySummary:             false,
				AlertSeverityThreshold:    "medium",
				NotifyOnNewAlert:          true,
				NotifyOnIncidentUpdate:    true,
				NotifyOnPlaybookExecution: false,
			}, nil
		}
		return nil, err
	}
	return &notifs, nil
}

// UpsertUserNotifications creates or updates user notification settings
func (r *ProfileRepository) UpsertUserNotifications(ctx context.Context, userID, orgID string, notifs *domain.UserNotifications) error {
	var id string
	err := r.Executor(ctx).QueryRowxContext(ctx, upsertUserNotificationsSQL,
		userID,
		sql.NullString{String: orgID, Valid: orgID != ""},
		notifs.EmailEnabled,
		notifs.PushEnabled,
		notifs.SlackEnabled,
		notifs.ProductUpdates,
		notifs.WeeklySummary,
		notifs.AlertSeverityThreshold,
		notifs.NotifyOnNewAlert,
		notifs.NotifyOnIncidentUpdate,
		notifs.NotifyOnPlaybookExecution,
	).Scan(&id)
	return err
}

type userSessionModel struct {
	ID               string         `db:"id"`
	UserID           string         `db:"user_id"`
	OrganizationID   sql.NullString `db:"organization_id"`
	SessionTokenHash string         `db:"session_token_hash"`
	IPAddress        sql.NullString `db:"ip_address"`
	UserAgent        sql.NullString `db:"user_agent"`
	Location         sql.NullString `db:"location"`
	IsRevoked        bool           `db:"is_revoked"`
	CreatedAt        time.Time      `db:"created_at"`
	LastActiveAt     time.Time      `db:"last_active_at"`
	ExpiresAt        time.Time      `db:"expires_at"`
}

func (m userSessionModel) ToDomain() domain.UserSession {
	var revokedAt sql.NullTime
	if m.IsRevoked {
		revokedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	return *domain.ReconstituteUserSession(
		m.ID,
		m.UserID,
		m.OrganizationID,
		m.SessionTokenHash,
		m.IPAddress,
		m.UserAgent,
		sql.NullString{},
		m.Location,
		!m.IsRevoked,
		m.LastActiveAt,
		m.ExpiresAt,
		m.CreatedAt,
		revokedAt,
	)
}

// GetUserSessions retrieves all active sessions for a user
func (r *ProfileRepository) GetUserSessions(ctx context.Context, userID, orgID string) ([]domain.UserSession, error) {
	var models []userSessionModel
	err := r.Executor(ctx).SelectContext(ctx, &models, getUserSessionsSQL, userID, sql.NullString{String: orgID, Valid: orgID != ""})
	if err != nil {
		return nil, err
	}
	sessions := make([]domain.UserSession, len(models))
	for i, m := range models {
		sessions[i] = m.ToDomain()
	}
	return sessions, nil
}

// GetUserSessionByID retrieves a specific session by ID
func (r *ProfileRepository) GetUserSessionByID(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	var model userSessionModel
	err := r.Executor(ctx).GetContext(ctx, &model, getUserSessionByIDSQL, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("SESSION NOT FOUND")
		}
		return nil, err
	}
	session := model.ToDomain()
	return &session, nil
}

// CreateUserSession creates a new user session
func (r *ProfileRepository) CreateUserSession(ctx context.Context, session *domain.UserSession) error {
	var id string
	err := r.Executor(ctx).QueryRowxContext(ctx, createUserSessionSQL,
		session.UserID(),
		session.OrganizationID(),
		session.SessionTokenHash(),
		session.IPAddress(),
		session.UserAgent(),
		session.Location(),
		session.ExpiresAt(),
	).Scan(&id)
	return err
}

// RevokeUserSession revokes a specific session for a user
func (r *ProfileRepository) RevokeUserSession(ctx context.Context, sessionID, userID string) error {
	result, err := r.Executor(ctx).ExecContext(ctx, revokeUserSessionSQL, sessionID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("SESSION NOT FOUND")
	}
	return nil
}

// UpdateSessionActivity updates the last_active_at timestamp
func (r *ProfileRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, updateSessionActivitySQL, sessionID)
	return err
}

// GetUserActivityLog retrieves paginated audit logs for a user
func (r *ProfileRepository) GetUserActivityLog(ctx context.Context, userID, orgID string, page, pageSize int) ([]domain.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	var total int
	err := r.Executor(ctx).GetContext(ctx, &total, countUserActivitySQL, userID, orgID)
	if err != nil {
		return nil, 0, err
	}

	var logs []domain.AuditLog
	err = r.Executor(ctx).SelectContext(ctx, &logs, getUserActivitySQL, userID, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CreateAuditLog creates a new audit log entry
func (r *ProfileRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	var id string
	err := r.Executor(ctx).QueryRowxContext(ctx, createAuditLogSQL,
		log.UserID,
		log.OrganizationID,
		log.ActionType,
		log.ResourceType,
		log.ResourceID,
		log.Metadata,
		log.IPAddress,
		log.UserAgent,
	).Scan(&id)
	if err == nil {
		log.ID = id
	}
	return err
}

// UpdateLastLogin updates the user's last login timestamp
func (r *ProfileRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, updateLastLoginSQL, userID)
	return err
}

// UpdateUserAvatar updates the user's avatar URL
func (r *ProfileRepository) UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, updateUserAvatarSQL, avatarURL, userID)
	return err
}
