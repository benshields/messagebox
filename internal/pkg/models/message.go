package models

import "time"

type ComposedMessage struct {
	Sender    string `json:"sender" binding:"required"`
	Recipient `json:"recipient" binding:"required"`
	Subject   string `json:"subject" binding:"required,min=1,max=255"`
	Body      string `json:"body" binding:"max=2000"`
}

type Recipient struct {
	Username  string `json:"username,omitempty"`
	Groupname string `json:"groupname,omitempty"`
}

type Message struct {
	Model       `binding:"required"`
	Re          int32  `json:"re,omitempty"`
	Sender      string `gorm:"-" json:"sender" binding:"required"`
	SenderID    int32  `gorm:"column:sender" json:"-"`
	Recipient   `gorm:"-" json:"recipient" binding:"required"`
	RecipientID int32     `gorm:"column:recipient" json:"-"`
	Subject     string    `json:"subject" binding:"required"`
	Body        string    `json:"body,omitempty"`
	SentAt      time.Time `gorm:"<-:create" json:"sentAt" binding:"required"`
}

type ReplyMessage struct {
	Re      int32
	Sender  string `json:"sender" binding:"required"`
	Subject string `json:"subject" binding:"required,min=1,max=255"`
	Body    string `json:"body,omitempty" binding:"max=2000"`
}
