package requests

type OnboardingRequest struct {
	FirstName   string  `json:"first_name" validate:"required"`
	LastName    string  `json:"last_name" validate:"required"`
	Email       string  `json:"email" validate:"email,required"`
	Password    string  `json:"password" validate:"password,required"`
	IndustryId  string  `json:"industry_id" validate:"required"`
	CompanyName string  `json:"company_name" validate:"required"`
	TimeZone    *string `json:"time_zone,omitempty"`
}

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"email,required"`
	Password  string `json:"password" validate:"password,required"`
}
type LoginRequest struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}
type GoogleAuthenticatorRequest struct {
	Code string `json:"code" validate:"required"`
}
type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"password,required"`
	ConfirmPassword string `json:"confirm_password" validate:"password,required"`
}
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"email,required"`
}
type VerifyResetTokenRequest struct {
	Token string `json:"token" validate:"required"`
}
type SendVerificationEmailRequest struct {
	Email string `json:"email" validate:"email,required"`
}
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
