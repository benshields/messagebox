package persistence

import (
	"fmt"
)

type UserRepository struct{}

var userRepository *UserRepository

func GetUserRepository() *UserRepository {
	if userRepository == nil {
		userRepository = &UserRepository{}
	}
	return userRepository
}

type User struct {
	ID   int32  `json:"id"`
	Name string `json:"username"`
}

// Mocks database persistence
var users = map[string]*User{}

var userID int32 = 3

func (r *UserRepository) Create(username string) (*User, error) {
	if _, ok := users[username]; ok {
		return nil, fmt.Errorf("username %s already exists", username)
	}
	user := &User{
		ID:   userID,
		Name: username,
	}
	userID++
	users[user.Name] = user
	return user, nil
}

func (r *UserRepository) Read(username string) (*User, error) {
	user, ok := users[username]
	if !ok {
		return nil, fmt.Errorf("username %s not found", username)
	}
	return user, nil
}
