package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/httperr"
	"github.com/benshields/messagebox/internal/pkg/models"
	"github.com/benshields/messagebox/internal/pkg/persistence"
)

type UserRegistration struct {
	Username string `json:"username" binding:"required,min=1,max=32"`
}

func CreateUser(c *gin.Context) {
	var req UserRegistration
	if err := c.BindJSON(&req); err != nil {
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	in := models.User{
		Name: req.Username,
	}

	r := persistence.GetUserRepository()
	out, err := r.Create(&in)
	if err != nil {
		httperr.NewError(c, http.StatusConflict, errors.New("user with the same username already registered"))
		return
	}

	resp := UserRegistration{Username: out.Name}
	c.JSON(http.StatusCreated, resp)
}

func GetUser(c *gin.Context) {
	req := c.Param("username")

	in := models.User{
		Name: req,
	}

	r := persistence.GetUserRepository()
	out, err := r.Read(&in)
	if err != nil {
		httperr.NewError(c, http.StatusNotFound, errors.New("user with given username does not exist"))
		return
	}

	c.JSON(http.StatusOK, out)
}

func GetMailbox(c *gin.Context) {
	var req models.UriUsername
	if err := c.BindUri(&req); err != nil {
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	in := models.User{
		Name: req.Username,
	}

	r := persistence.GetUserRepository()
	out, err := r.GetMailbox(&in)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			httperr.NewError(c, http.StatusNotFound, errors.New("user with given username does not exist"))
			return
		default:
			httperr.NewError(c, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	}

	c.JSON(http.StatusOK, out)
}
