package persistence

import (
	"github.com/benshields/messagebox/internal/pkg/db"
	models "github.com/benshields/messagebox/internal/pkg/models/users"
)

type UserRepository struct{}

var userRepository *UserRepository

func GetUserRepository() *UserRepository {
	if userRepository == nil {
		userRepository = &UserRepository{}
	}
	return userRepository
}

func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	result := db.Get().Create(&user)
	return user, result.Error
}

func (r *UserRepository) Read(user *models.User) (*models.User, error) {
	result := db.Get().First(&user)
	return user, result.Error
}
