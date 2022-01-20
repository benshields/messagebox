package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/pkg/httperr"
	"github.com/benshields/messagebox/internal/pkg/models"
	"github.com/benshields/messagebox/internal/pkg/persistence"
)

type UserRegistration struct {
	Username string `json:"username" binding:"required"`
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
	if err != nil { // TODO switch on gorm errors here
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
	if err != nil { // TODO switch on gorm errors here
		httperr.NewError(c, http.StatusNotFound, errors.New("user with given username does not exist"))
		return
	}

	c.JSON(http.StatusOK, out)
}
