package router

import (
	"github.com/gin-gonic/gin"

	"github.com/benshields/messagebox/internal/api/controllers"
	"github.com/benshields/messagebox/internal/api/middleware"
)

func Setup() *gin.Engine {
	r := gin.New()

	r.NoRoute(middleware.NoRouteHandler())
	r.NoMethod(middleware.NoMethodHandler())
	r.Use(gin.Recovery())

	r.POST("/users", controllers.CreateUser)
	r.GET("/users/:username", controllers.GetUser)
	r.GET("users/:username/mailbox", controllers.GetMailbox)

	r.POST("/groups", controllers.CreateGroup)
	r.GET("/groups/:groupname", controllers.GetGroup)

	r.POST("/messages", controllers.CreateMessage)
	r.GET("/messages/:id", controllers.GetMessage)
	r.POST("/messages/:id/replies", controllers.CreateReply)
	r.GET("/messages/:id/replies", controllers.GetReplies)

	return r
}
