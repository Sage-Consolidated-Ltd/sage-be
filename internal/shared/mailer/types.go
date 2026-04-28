package mailer

type MemberInvitationEmailData struct {
	OrganizationName string
	Role             string
	InviteLink       string
}

type VerificationEmailData struct {
	Name      string
	OTP       string
	ExpiresIn string
}
