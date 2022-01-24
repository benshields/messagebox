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

func (r *UserRepository) GetByID(user *models.User) (*models.User, error) {
	result := db.Get().Take(&user, "id = ?", user.ID)
	return user, result.Error
}

func (r *UserRepository) GetMailbox(user *models.User) ([]*models.Message, error) { // TODO this must order by created at
	var err error
	messages := make([]*models.Message, 0)
	err = db.Get().Transaction(func(tx *gorm.DB) error {
		user, err = r.Read(user)
		if err != nil {
			return err
		}

		userGroups, err := GetGroupRepository().FindByUserID(user)
		if err != nil {
			return err
		}

		ids := make([]int32, len(userGroups)+1)
		for i, userGroup := range userGroups {
			ids[i] = userGroup.GroupID
		}
		ids[len(userGroups)] = user.ID

		messages, err = GetMessageRepository().FindByRecipientID(ids)
		if err != nil {
			return err
		}

		return nil
	})
	return messages, err
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

func (r *GroupRepository) GetByID(group *models.Group) (*models.Group, error) {
	result := db.Get().Take(&group, "id = ?", group.ID)
	return group, result.Error
}

func (r *GroupRepository) FindByUserID(user *models.User) ([]*models.UserGroup, error) {
	var userGroups []*models.UserGroup
	result := db.Get().Find(&userGroups, "user_id = ?", user.ID)
	return userGroups, result.Error
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
		senderOut, err := GetUserRepository().Read(senderIn) // TODO link the transactions together
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

func (r *MessageRepository) Read(message *models.Message) (*models.Message, error) {
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&message, "id = ?", message.ID).Error; err != nil {
			return err
		}

		senderIn := &models.User{
			Model: models.Model{
				ID: message.SenderID,
			},
		}
		senderOut, err := GetUserRepository().GetByID(senderIn)
		if err != nil {
			return err
		}
		message.Sender = senderOut.Name

		if message.RecipientID > 0 {
			userRecipientIn := &models.User{
				Model: models.Model{
					ID: message.RecipientID,
				},
			}
			userRecipientOut, err := GetUserRepository().GetByID(userRecipientIn)
			if err != nil {
				return err
			}
			message.Recipient = models.Recipient{
				Username: userRecipientOut.Name,
			}
		} else if message.RecipientID < 0 {
			groupRecipientIn := &models.Group{
				Model: models.Model{
					ID: message.RecipientID,
				},
			}
			groupRecipientOut, err := GetGroupRepository().GetByID(groupRecipientIn)
			if err != nil {
				return err
			}
			message.Recipient = models.Recipient{
				Groupname: groupRecipientOut.Name,
			}
		} else {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	return message, err
}

func (r *MessageRepository) CreateReply(message *models.Message) (*models.Message, error) {
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		originalMessageIn := &models.Message{
			Model: models.Model{
				ID: message.Re,
			},
		}
		originalMessageOut, err := r.Read(originalMessageIn)
		if err != nil {
			return err
		}

		senderIn := &models.User{
			Name: message.Sender,
		}
		senderOut, err := GetUserRepository().Read(senderIn)
		if err != nil {
			return err
		}
		message.SenderID = senderOut.ID

		if originalMessageOut.RecipientID > 0 { // TODO everywhere I check the sign of the ID I should be using a util func to test if this is user or group ID
			message.Recipient = models.Recipient{
				Username: originalMessageOut.Sender,
			}
			message.RecipientID = originalMessageOut.SenderID
		} else {
			message.Recipient = models.Recipient{
				Groupname: originalMessageOut.Recipient.Groupname,
			}
			message.RecipientID = originalMessageOut.RecipientID
		}

		message.SentAt = time.Now().UTC()

		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		return nil
	})

	return message, err
}

func (r *MessageRepository) GetReplies(message *models.Message) ([]*models.Message, error) { // TODO this must order by created at
	var replies []*models.Message
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&message, "id = ?", message.ID).Error; err != nil {
			return err
		}

		if err := tx.Find(&replies, "re = ?", message.ID).Error; err != nil {
			return err
		}

		for i, in := range replies {
			out, err := r.Read(in)
			if err != nil {
				return err
			}
			replies[i] = out
		}

		return nil
	})

	return replies, err
}

func (r *MessageRepository) FindByRecipientID(ids []int32) ([]*models.Message, error) {
	var messages []*models.Message
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("sent_at").Find(&messages, "recipient IN ?", ids).Error; err != nil {
			return err
		}

		for i, in := range messages {
			out, err := r.Read(in)
			if err != nil {
				return err
			}
			messages[i] = out
		}

		return nil
	})

	return messages, err
}
