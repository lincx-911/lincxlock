package route

import (
	"github.com/lincx-911/lincxlock/httpserver/api"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine{
	router := gin.Default()
	rLock := router.Group("/lincxlock")
	{
		rLock.POST("/lock",api.Lock)
		rLock.POST("/unlock",api.Unlock)
	}
	return router
}