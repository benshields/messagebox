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

func CreateMessage(c *gin.Context) {
	var req models.ComposedMessage
	if err := c.BindJSON(&req); err != nil {
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	// ensure only 1 recipient
	switch {
	case req.Recipient.Username == "" && req.Recipient.Groupname == "":
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	case req.Recipient.Username != "" && req.Recipient.Groupname != "":
		httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	r := persistence.GetMessageRepository()
	out, err := r.Create(&req)
	if err != nil {
		switch { // TODO improve error handling accross the entire API
		case errors.Is(err, gorm.ErrRecordNotFound):
			httperr.NewError(c, http.StatusNotFound, errors.New("sender or recipient does not exist")) // TODO improve specificity
			return
		case errors.Is(err, gorm.ErrInvalidValue), errors.Is(err, gorm.ErrInvalidValueOfLength):
			httperr.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
			return
		default:
			httperr.NewError(c, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	}

	c.JSON(http.StatusCreated, out)
}
