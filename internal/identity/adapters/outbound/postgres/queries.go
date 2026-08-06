package postgres

const (
	CREATE_USER = `
	INSERT INTO users (
		first_name,
		last_name,
		email,
		time_zone,
		password_hash
	) VALUES ($1, $2, $3, $4, $5)
	RETURNING *;
	`

	GET_USER_BY_EMAIL = `
	SELECT * FROM users WHERE email = $1
	`

	GET_USER_BY_ID = `
	SELECT * FROM users WHERE id = $1
	`

	MARK_EMAIL_VERIFIED = `
	UPDATE users SET is_verified = true WHERE email = $1
	`

	CREATE_ORGANIZATION = `
	INSERT INTO organizations (
		name,
		owner_id,
		industry_id
	) VALUES ($1, $2, $3)
	RETURNING id;
	`

	ADD_USER_TO_ORGANIZATION = `
	INSERT INTO organization_members (
		organization_id,
		user_id,
		role_id, 
		status
	) VALUES ($1, $2, $3, $4)
	`

	GET_USER_ORGANIZATIONS = `
	SELECT 
		o.id, 
		o.name, 
		o.owner_id, 
		COALESCE(i.name, '') AS industry,
		COALESCE(r.name, '') AS role, 
		om.status,
		o.created_at, 
		o.updated_at, 
		o.deleted_at 
	FROM organizations o
	JOIN organization_members om ON o.id = om.organization_id
	LEFT JOIN industries i ON o.industry_id = i.id
	LEFT JOIN organization_roles r ON om.role_id = r.id
	WHERE om.user_id = $1
	`

	ENABLE_2FA = `
	UPDATE users 
	SET 
		two_factor_enabled = true,
		two_factor_secret = $1
	WHERE id = $2
	`

	GET_TOTP_SECRET = `
	SELECT two_factor_secret FROM users 
	WHERE id = $1 
	AND two_factor_enabled = true
	AND is_verified = true
	`

	UPDATE_USER_PASSWORD = `
	UPDATE users SET password_hash = $1 WHERE email = $2
	`

	UPDATE_USER = `
	UPDATE users SET 
		first_name = $1,
		last_name = $2,
		time_zone = $3,
		updated_at = NOW()
	WHERE id = $4
	`

	UPDATE_USER_CONTACT_INFO = `
	UPDATE users SET 
		phone_number = CASE WHEN $1 != '' THEN $1 ELSE phone_number END,
		backup_email = CASE WHEN $2 != '' THEN $2 ELSE backup_email END,
		updated_at = NOW()
	WHERE id = $3
	`

	GET_ORGANIZATION_ROLE_ID = `
	SELECT id FROM organization_roles WHERE name = $1
	`

	INVITE_MEMBER_TO_ORGANIZATION = `
	INSERT INTO organization_invites (
		organization_id,
		email,
		role_id,
		invited_by,
		token_hash,
		expires_at
	)
	SELECT $1, LOWER($2), $3, $4, $5, $6
	WHERE NOT EXISTS (
		SELECT 1 FROM organization_invites oi
		WHERE oi.organization_id = $1
		AND LOWER(oi.email) = LOWER($2)
		AND oi.status = 'pending'
	)
	ON CONFLICT DO NOTHING
	RETURNING id;
	`
)
