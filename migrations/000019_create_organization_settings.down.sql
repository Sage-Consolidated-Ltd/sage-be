-- Drop organization_settings table
DROP TABLE IF EXISTS organization_settings;
DROP TRIGGER IF EXISTS trigger_org_settings_updated_at ON organization_settings;
DROP FUNCTION IF EXISTS update_org_settings_updated_at();
