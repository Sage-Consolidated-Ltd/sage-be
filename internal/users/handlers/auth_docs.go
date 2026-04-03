package handlers

import (
	_ "sage-backend/internal/users/requests"
	_ "sage-backend/internal/users/models"
)

// @Summary Create User
// @Description Creates a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "User Details"
// @Success 201
// @Router /auth/register [post]
func _CreateUser(){}

// @Summary      Login User with OAUTH
// @Description  Logs in a user using OAUTH (Google, Github)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        provider  path      string  true  "OAuth Provider (e.g. google, github)"
// @Success      200
// @Router       /auth/login/{provider} [get]
func _BeginAuthLogin(){}

// @Summary      Login User
// @Description  Logs in a user using email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.LoginRequest true "Login Credentials"
// @Success      200       {object}  models.GetUserResponse
// @Router       /auth/login [post]
func _Login(){}