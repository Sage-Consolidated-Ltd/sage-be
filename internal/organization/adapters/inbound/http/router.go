package http

import (
	"sage-backend/internal/shared/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetUpRouter(router fiber.Router, ch *CompanyHandler, m *middlewares.AuthMiddleware) {
	RegisterCompanyRoutes(router, ch, m)
}

func RegisterCompanyRoutes(router fiber.Router, ch *CompanyHandler, m *middlewares.AuthMiddleware) {
	company := router.Group("/company")
	company.Get("/industries", ch.GetIndustries)
	company.Get("/organization-roles", ch.GetOrganizationRoles)
	company.Post("/invite", m.RequireAuth, ch.InviteMember)
	company.Get("/invitations/accept", m.RequireAuth, ch.AcceptInvitation)

	org := router.Group("/organization", m.RequireAuth)
	org.Get("/", ch.GetOrganization)
	org.Patch("/", ch.UpdateOrganization)
	org.Get("/profile", ch.GetOrganizationProfile)
	org.Patch("/profile/details", ch.UpdateCompanyDetails)
	org.Patch("/branding", ch.UpdateOrganizationBranding)
	org.Post("/branding/upload", middlewares.ValidateFileUpload("logo", middlewares.MaxBannerSize, middlewares.AvatarMimeTypes), ch.UploadBrandingLogo)

	members := org.Group("/members")
	members.Get("/", ch.ListMembers)
	members.Post("/invite", ch.InviteMembers)
	members.Patch("/:id/role", ch.UpdateMemberRole)
	members.Post("/:id/reset-mfa", ch.ResetMemberMFA)
	members.Patch("/:id/status", ch.UpdateMemberStatus)
	members.Delete("/:id", ch.RemoveMember)

	settings := org.Group("/settings")
	settings.Get("/", ch.GetOrganizationSettings)
	settings.Patch("/", ch.UpdateOrganizationSettings)

	org.Get("/permissions", ch.ListPermissions)
	org.Get("/permission-groups", ch.ListPermissionGroups)

	customRoles := org.Group("/custom-roles")
	customRoles.Get("/", ch.ListCustomRoles)
	customRoles.Post("/", ch.CreateCustomRole)
	customRoles.Get("/:id", ch.GetCustomRole)
	customRoles.Patch("/:id", ch.UpdateCustomRole)
	customRoles.Delete("/:id", ch.DeleteCustomRole)
}
