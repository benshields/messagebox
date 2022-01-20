package controllers

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/pkg/httperr"
	"github.com/benshields/messagebox/internal/pkg/models"
	"github.com/benshields/messagebox/internal/pkg/persistence"
)

type GroupCreation struct {
	Groupname string   `json:"groupname" binding:"required"`
	Usernames []string `json:"usernames" binding:"required"`
}

func CreateGroup(c *gin.Context) {
	var req GroupCreation
	if err := c.BindJSON(&req); err != nil {
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	in := models.Group{
		Name:  req.Groupname,
		Users: make([]models.User, len(req.Usernames)),
	}
	for i, uname := range req.Usernames {
		u := models.User{
			Name: uname,
		}
		in.Users[i] = u
	}

	r := persistence.GetGroupRepository()
	out, err := r.Create(&in)
	if err != nil {
		switch { // TODO improve error handling accross the entire API
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			httperr.NewError(c, http.StatusConflict, errors.New("group with the same groupname already registered"))
			return
		case errors.Is(err, gorm.ErrRecordNotFound):
			httperr.NewError(c, http.StatusNotFound, errors.New("one or more users with given usernames do not exist"))
			return
		case errors.Is(err, gorm.ErrInvalidValue), errors.Is(err, gorm.ErrInvalidValueOfLength):
			httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
			return
		default:
			httperr.NewError(c, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	}

	resp := GroupCreation{
		Groupname: out.Name,
		Usernames: make([]string, len(out.Users)),
	}
	for i, u := range out.Users {
		resp.Usernames[i] = u.Name
	}
	c.JSON(http.StatusCreated, resp)
}

func GetGroup(c *gin.Context) {
	req := c.Param("groupname")

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
