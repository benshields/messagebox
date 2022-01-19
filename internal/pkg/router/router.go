package router

import (
	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/api/controllers"
	"github.com/benshields/messagebox/internal/api/middleware"
)

func Setup() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.NoRoute(middleware.NoRouteHandler())
	r.NoMethod(middleware.NoMethodHandler())

	r.POST("/users", controllers.CreateUser)
	r.GET("/users/:username", controllers.GetUser)

	return r
}
