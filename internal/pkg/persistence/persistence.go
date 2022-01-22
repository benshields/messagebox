package persistence

import (
	"time"

	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/db"
	"github.com/benshields/messagebox/internal/pkg/models"
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
	result := db.Get().Take(&user, "name = ?", user.Name)
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
		tx = tx.Where("name IN ?", unames).Find(&group.Users) // TODO I think this allows SQL injection attack
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
	result := db.Get().Take(&group, "name = ?", group.Name)
	return group, result.Error
}

////////// TODO split to another file?

type MessageRepository struct{}

var messageRepository *MessageRepository

func GetMessageRepository() *MessageRepository {
	if messageRepository == nil {
		messageRepository = &MessageRepository{}
	}
	return messageRepository
}

func (r *MessageRepository) Create(composedMsg *models.ComposedMessage) (*models.Message, error) {
	msg := &models.Message{
		Sender:    composedMsg.Sender,
		Recipient: composedMsg.Recipient,
		Subject:   composedMsg.Subject,
		Body:      composedMsg.Body,
		SentAt:    time.Now().UTC(),
	}
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		var err error

		// ensure sender exists
		senderIn := &models.User{
			Name: composedMsg.Sender,
		}
		senderOut, err := GetUserRepository().Read(senderIn)
		if err != nil {
			return err
		}
		msg.SenderID = senderOut.ID

		// ensure recipient exists
		if composedMsg.Recipient.Username != "" {
			userRecipientIn := &models.User{
				Name: composedMsg.Recipient.Username,
			}
			userRecipientOut, err := GetUserRepository().Read(userRecipientIn)
			if err != nil {
				return err
			}
			msg.RecipientID = userRecipientOut.ID
		} else {
			groupRecipientIn := &models.Group{
				Name: composedMsg.Recipient.Groupname,
			}
			groupRecipientOut, err := GetGroupRepository().Read(groupRecipientIn)
			if err != nil {
				return err
			}
			msg.RecipientID = groupRecipientOut.ID
		}

		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		return nil
	})

	return msg, err
}
