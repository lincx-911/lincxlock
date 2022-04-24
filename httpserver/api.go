package httpserver

import (
	"github.com/lincx-911/lincxlock/httpserver/route"
	"strconv"

	"github.com/gin-gonic/gin"
)

var Route *gin.Engine

func init() {
	Route = route.InitRouter()
}

func StartHttp(port int)error{
	return Route.Run(":"+strconv.Itoa(port))
}

func StartHttps(port int)error{
	return Route.RunTLS(":"+strconv.Itoa(port),"./server.crt","server.key")
}