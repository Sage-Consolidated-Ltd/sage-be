DROP INDEX idx_organization_industries_org_industry;
DROP TABLE organization_industries;
ALTER TABLE organizations ADD COLUMN industry_id UUID REFERENCES industries(id) ON DELETE SET NULL;
CREATE INDEX idx_organizations_industry_id ON organizations(industry_id) WHERE deleted_at IS NULL
