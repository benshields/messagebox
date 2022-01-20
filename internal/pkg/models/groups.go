package models

type Group struct {
	Model
	Name  string `gorm:"not null" json:"groupname"`
	Users []User `gorm:"many2many:user_groups;"`
}
