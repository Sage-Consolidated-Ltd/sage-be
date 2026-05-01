INSERT INTO roles (name, description) VALUES
  ('admin', 'Administrator with full access to all resources and settings.'),
  ('user', 'Regular user with access to standard features and functionalities.'),
  ('support', 'Support staff with access to customer support tools and resources.'),
  ('super_admin', 'Super administrator with elevated privileges for managing the system.')
  ON CONFLICT (name) DO NOTHING;

  INSERT INTO organization_roles (name, description)
  VALUES
    ('owner',   'Full control over the organization'),
    ('admin',   'Manages users, settings, and resources'),
    ('analyst', 'Can view and analyze data'),
    ('member',  'Standard access to organization resources'),
    ('viewer',  'Read-only access'),
    ('automation_admin', 'Responsible for managing automated processes'),
    ('billing_admin', 'Responsible for managing billing and financial operations')
  ON CONFLICT (name) DO NOTHING;