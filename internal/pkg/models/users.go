package models

type User struct {
	Model
	Name string `gorm:"not null" json:"username"`
}
