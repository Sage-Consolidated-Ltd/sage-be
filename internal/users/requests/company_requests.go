package requests

type InviteMemberRequest struct {
	Email  string `json:"email" validate:"email,required"`
	RoleId string `json:"role_id" validate:"required"`
}

type BulkInviteMembersRequest struct {
	Invites []InviteMemberRequest `json:"invites" validate:"required,dive"`
}
