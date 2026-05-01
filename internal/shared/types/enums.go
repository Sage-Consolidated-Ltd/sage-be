package types

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case AdminRole, UserRole:
		return true
	default:
		return false
	}
}

// Severity levels for alerts and events
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

func (s Severity) IsValid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// EventCategory categorizes security events
type EventCategory string

const (
	EventCategoryAuthentication EventCategory = "authentication"
	EventCategoryAccess         EventCategory = "access"
	EventCategoryNetwork        EventCategory = "network"
	EventCategorySystem         EventCategory = "system"
	EventCategoryApplication    EventCategory = "application"
	EventCategoryData           EventCategory = "data"
)

func (e EventCategory) IsValid() bool {
	switch e {
	case EventCategoryAuthentication, EventCategoryAccess, EventCategoryNetwork,
		EventCategorySystem, EventCategoryApplication, EventCategoryData:
		return true
	default:
		return false
	}
}

// ParseStatus indicates parsing outcome for an event
type ParseStatus string

const (
	ParseStatusPending ParseStatus = "pending"
	ParseStatusSuccess ParseStatus = "success"
	ParseStatusFailed  ParseStatus = "failed"
	ParseStatusPartial ParseStatus = "partial"
)

func (p ParseStatus) IsValid() bool {
	switch p {
	case ParseStatusPending, ParseStatusSuccess, ParseStatusFailed, ParseStatusPartial:
		return true
	default:
		return false
	}
}

// QualityStatus represents data source quality health
type QualityStatus string

const (
	QualityGood    QualityStatus = "good"
	QualityWarning QualityStatus = "warning"
	QualityPartial QualityStatus = "partial"
	QualityError   QualityStatus = "error"
)

func (q QualityStatus) IsValid() bool {
	switch q {
	case QualityGood, QualityWarning, QualityPartial, QualityError:
		return true
	default:
		return false
	}
}

// ParserType defines parser technology
type ParserType string

const (
	ParserTypeRegex    ParserType = "regex"
	ParserTypeJSON     ParserType = "json"
	ParserTypeCSV      ParserType = "csv"
	ParserTypeKeyValue ParserType = "key_value"
	ParserTypeAINLP    ParserType = "ai_nlp"
)

func (p ParserType) IsValid() bool {
	switch p {
	case ParserTypeRegex, ParserTypeJSON, ParserTypeCSV, ParserTypeKeyValue, ParserTypeAINLP:
		return true
	default:
		return false
	}
}

// ParserStatus indicates parser health
type ParserStatus string

const (
	ParserStatusActive   ParserStatus = "active"
	ParserStatusWarning  ParserStatus = "warning"
	ParserStatusError    ParserStatus = "error"
	ParserStatusDisabled ParserStatus = "disabled"
)

func (p ParserStatus) IsValid() bool {
	switch p {
	case ParserStatusActive, ParserStatusWarning, ParserStatusError, ParserStatusDisabled:
		return true
	default:
		return false
	}
}

// JobStatus for async jobs
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

func (j JobStatus) IsValid() bool {
	switch j {
	case JobStatusQueued, JobStatusRunning, JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	default:
		return false
	}
}

// AlertStatus for incident lifecycle (from AGENTS.md)
type AlertStatus string

const (
	AlertStatusNew             AlertStatus = "new"
	AlertStatusInvestigating   AlertStatus = "investigating"
	AlertStatusContained       AlertStatus = "contained"
	AlertStatusResolved        AlertStatus = "resolved"
	AlertStatusDismissed       AlertStatus = "dismissed"
	AlertStatusPendingApproval AlertStatus = "pending_approval"
	AlertStatusNeedsReview     AlertStatus = "needs_review"
)

func (a AlertStatus) IsValid() bool {
	switch a {
	case AlertStatusNew, AlertStatusInvestigating, AlertStatusContained,
		AlertStatusResolved, AlertStatusDismissed, AlertStatusPendingApproval,
		AlertStatusNeedsReview:
		return true
	default:
		return false
	}
}

// RuleType detection method
type RuleType string

const (
	RuleTypeBehavioral RuleType = "behavioral"
	RuleTypeRuleBased  RuleType = "rule_based"
	RuleTypeAIDriven   RuleType = "ai_driven"
)

func (r RuleType) IsValid() bool {
	switch r {
	case RuleTypeBehavioral, RuleTypeRuleBased, RuleTypeAIDriven:
		return true
	default:
		return false
	}
}

// RuleSource indicates rule origin
type RuleSource string

const (
	RuleSourceCustom      RuleSource = "custom"
	RuleSourcePrebuilt    RuleSource = "prebuilt"
	RuleSourceAIGenerated RuleSource = "ai_generated"
)

func (r RuleSource) IsValid() bool {
	switch r {
	case RuleSourceCustom, RuleSourcePrebuilt, RuleSourceAIGenerated:
		return true
	default:
		return false
	}
}

// ActionType from AGENTS.md
type ActionType string

const (
	ActionBlockIP            ActionType = "block_ip"
	ActionDisableUserAccount ActionType = "disable_user_account"
	ActionRevokeSessions     ActionType = "revoke_sessions"
	ActionRemoveForwarding   ActionType = "remove_forwarding_rule"
	ActionQuarantineDevice   ActionType = "quarantine_device"
	ActionContainAsset       ActionType = "contain_asset"
	ActionNotifyAdmin        ActionType = "notify_admin"
	ActionCreateTicket       ActionType = "create_ticket"
)

func (a ActionType) IsValid() bool {
	switch a {
	case ActionBlockIP, ActionDisableUserAccount, ActionRevokeSessions,
		ActionRemoveForwarding, ActionQuarantineDevice, ActionContainAsset,
		ActionNotifyAdmin, ActionCreateTicket:
		return true
	default:
		return false
	}
}

// ActionExecStatus from AGENTS.md
type ActionExecStatus string

const (
	ActionStatusPending          ActionExecStatus = "pending"
	ActionStatusRunning          ActionExecStatus = "running"
	ActionStatusSuccess          ActionExecStatus = "success"
	ActionStatusFailed           ActionExecStatus = "failed"
	ActionStatusCancelled        ActionExecStatus = "cancelled"
	ActionStatusRequiresApproval ActionExecStatus = "requires_approval"
)

func (a ActionExecStatus) IsValid() bool {
	switch a {
	case ActionStatusPending, ActionStatusRunning, ActionStatusSuccess,
		ActionStatusFailed, ActionStatusCancelled, ActionStatusRequiresApproval:
		return true
	default:
		return false
	}
}

// AnalystDecision from AGENTS.md
type AnalystDecision string

const (
	DecisionResolved      AnalystDecision = "resolved"
	DecisionContained     AnalystDecision = "contained"
	DecisionInvestigating AnalystDecision = "investigating"
	DecisionFalsePositive AnalystDecision = "false_positive"
	DecisionDismissed     AnalystDecision = "dismissed"
	DecisionNeedsReview   AnalystDecision = "needs_review"
)

func (a AnalystDecision) IsValid() bool {
	switch a {
	case DecisionResolved, DecisionContained, DecisionInvestigating,
		DecisionFalsePositive, DecisionDismissed, DecisionNeedsReview:
		return true
	default:
		return false
	}
}
