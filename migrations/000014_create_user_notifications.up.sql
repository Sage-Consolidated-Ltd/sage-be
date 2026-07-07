-- User notification preferences table
CREATE TABLE IF NOT EXISTS user_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    push_enabled BOOLEAN DEFAULT false,
    slack_enabled BOOLEAN DEFAULT false,
    alert_severity_threshold VARCHAR(20) DEFAULT 'medium' CHECK (alert_severity_threshold IN ('low', 'medium', 'high', 'critical')),
    notify_on_new_alert BOOLEAN DEFAULT true,
    notify_on_incident_update BOOLEAN DEFAULT true,
    notify_on_playbook_execution BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, organization_id)
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_notifications_org_id ON user_notifications(organization_id);
