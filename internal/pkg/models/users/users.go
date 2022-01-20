package users

import (
	"github.com/benshields/messagebox/internal/pkg/models"
)

type User struct {
	models.Model
	Name string `gorm:"not null" json:"username"`
}
