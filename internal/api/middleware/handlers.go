package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/pkg/httperr"
)

func NoMethodHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		httperr.NewError(c, http.StatusMethodNotAllowed, errors.New("method not permitted"))
	}
}

func NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		httperr.NewError(c, http.StatusNotFound, errors.New("no route found"))
	}
}
