package mailer

var (
	MEMBER_INVITE_TEMPLATE = `
		<!DOCTYPE html>
			<html lang="en">
				<body style="margin:0; padding:0; background-color:#f4f4f7; font-family:Arial, sans-serif;">
		
					<div style="max-width:600px; margin:40px auto; background:#ffffff; border-radius:8px; padding:32px;">
			
					<div style="font-size:20px; font-weight:bold; margin-bottom:16px;">
						You're invited to join {{.OrganizationName}}
					</div>

					<div style="font-size:14px; color:#555; line-height:1.6;">
			

					<p style="margin:0 0 16px 0;">
						You’ve been invited to join <strong>{{.OrganizationName}}</strong> as a/an {{.Role}}.
						Click the button below to accept your invitation and get started.
					</p>

					<a href="{{.InviteLink}}" 
					style="display:inline-block; margin-top:16px; padding:12px 20px; background-color:#4CAF50; color:#ffffff; text-decoration:none; border-radius:6px; font-weight:bold;">
						Accept Invitation
					</a>

					<p style="margin:24px 0 8px 0; word-break:break-all;">
						If the button doesn't work, copy and paste this link into your browser:
					</p>

					<p style="margin:0 0 16px 0; word-break:break-all;">
						<a href="{{.InviteLink}}" style="color:#4CAF50;">{{.InviteLink}}</a>
					</p>

					<p style="margin:0 0 16px 0;">
						If you weren’t expecting this invite, you can safely ignore this email.
					</p>

				</div>

			</div>

			</body>
			</html>
	`
	VERIFICATION_EMAIL_TEMPLATE = `
    <!DOCTYPE html>
    <html lang="en">
        <body style="margin:0; padding:0; background-color:#f4f4f7; font-family:Arial, sans-serif;">

            <div style="max-width:600px; margin:40px auto; background:#ffffff; border-radius:8px; padding:32px;">

                <div style="font-size:20px; font-weight:bold; margin-bottom:16px;">
                    Verify your email address
                </div>

                <div style="font-size:14px; color:#555; line-height:1.6;">

                    <p style="margin:0 0 16px 0;">
                        Hi <strong>{{.Name}}</strong>, use the verification code below to confirm your email address.
                        This code expires in <strong>{{.ExpiresIn}}</strong>.
                    </p>

                    <div style="margin:24px 0; text-align:center;">
                        <div style="display:inline-block; background-color:#f4f4f7; border-radius:8px; padding:16px 32px;">
                            <span style="font-size:32px; font-weight:bold; letter-spacing:8px; color:#222222;">
                                {{.OTP}}
                            </span>
                        </div>
                    </div>

                    <p style="margin:0 0 16px 0;">
                        Enter this code on the verification page to activate your account.
                    </p>

                    <p style="margin:0 0 16px 0; font-size:13px; color:#888;">
                        If you didn't request this, you can safely ignore this email.
                        Someone may have entered your email address by mistake.
                    </p>

                </div>

            </div>

        </body>
    </html>
`
)