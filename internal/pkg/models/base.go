package models

type Model struct {
	ID int32 `gorm:"column:id;primary_key;" json:"id"`
}

type UriId struct {
	ID int32 `uri:"id" binding:"required,numeric"`
}

type UriUsername struct {
	Username string `uri:"username" binding:"required"`
}
