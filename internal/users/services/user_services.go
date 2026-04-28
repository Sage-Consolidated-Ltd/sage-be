package services

import (
	"sage-backend/internal/users/repositories"

	"github.com/redis/go-redis/v9"
)

type UserServicesInt interface {
}
type UserServices struct {
	userRepo    repositories.UsersRepositoryInt
	redisClient *redis.Client
}

func NewUsersServices(usersRepo repositories.UsersRepositoryInt, redisClient *redis.Client) UserServicesInt {
	return &UserServices{
		userRepo:    usersRepo,
		redisClient: redisClient,
	}
}
