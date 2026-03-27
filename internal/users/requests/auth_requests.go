package requests

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"email,required"`
	Password  string `json:"password" validate:"password,required"`
}
