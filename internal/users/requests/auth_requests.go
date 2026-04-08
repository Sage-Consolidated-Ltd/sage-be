package requests

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
	Password string `json:"password" validate:"password,required"`
	ConfirmPassword string `json:"confirm_password" validate:"password,required"`
}
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"email,required"`
}
type VerifyResetTokenRequest struct {
	Token string `json:"token" validate:"required"`
}