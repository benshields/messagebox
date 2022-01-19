package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/pkg/httperr"
	"github.com/benshields/messagebox/internal/pkg/persistence"
)

type UserRegistration struct {
	Username string `json:"username" binding:"required"`
}

func CreateUser(c *gin.Context) {
	r := persistence.GetUserRepository()
	var userInput UserRegistration
	if err := c.BindJSON(&userInput); err != nil {
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}
	user, err := r.Create(userInput.Username)
	if err != nil {
		httperr.NewError(c, http.StatusConflict, errors.New("user with the same username already registered"))
		return
	}
	resp := UserRegistration{Username: user.Name}
	c.JSON(http.StatusCreated, resp)
}

func GetUser(c *gin.Context) {
	r := persistence.GetUserRepository()
	username := c.Param("username")
	user, err := r.Read(username)
	if err != nil {
		httperr.NewError(c, http.StatusNotFound, errors.New("user with given username does not exist"))
		return
	}
	c.JSON(http.StatusOK, user)
}
