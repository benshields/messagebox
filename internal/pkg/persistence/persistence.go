package persistence

import (
	"github.com/benshields/messagebox/internal/pkg/db"
	"github.com/benshields/messagebox/internal/pkg/models"
	"gorm.io/gorm"
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

////////// TODO split to another file?

type GroupRepository struct{}

var groupRepository *GroupRepository

func GetGroupRepository() *GroupRepository {
	if groupRepository == nil {
		groupRepository = &GroupRepository{}
	}
	return groupRepository
}

func (r *GroupRepository) Create(group *models.Group) (*models.Group, error) {
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		// first ensure all users exist
		unames := make([]string, len(group.Users))
		for i, u := range group.Users {
			unames[i] = u.Name
		}
		tx = tx.Where("name IN ?", unames).Find(&group.Users)
		if err := tx.Error; err != nil {
			return err
		}
		if tx.RowsAffected != int64(len(unames)) {
			return gorm.ErrRecordNotFound
		}

		// create group and user_group entries but leave existing users table alone
		if err := tx.Omit("Users.*").Create(&group).Error; err != nil {
			return err
		}

		return nil
	})

	return group, err
}

func (r *GroupRepository) Read(group *models.Group) (*models.Group, error) {
	result := db.Get().First(&group)
	return group, result.Error
}
