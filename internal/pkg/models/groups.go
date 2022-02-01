package models

type Group struct {
	Model
	Name  string `gorm:"not null" json:"groupname"`
	Users []User `gorm:"-"`
}

type UserGroup struct {
	Model
	GroupID int32 `gorm:"column:group_id"`
	UserID  int32 `gorm:"column:user_id"`
}
