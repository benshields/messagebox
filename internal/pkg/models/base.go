package models

type Model struct {
	ID int32 `gorm:"column:id;primary_key;" json:"id"`
}
