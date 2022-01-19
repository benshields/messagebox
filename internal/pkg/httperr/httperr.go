package httperr

import "github.com/gin-gonic/gin"

func NewError(c *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	c.JSON(status, er)
}

type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}
