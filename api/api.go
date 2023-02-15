package api

import (
	"github.com/cybersamx/teapot/httpx"
	"github.com/gin-gonic/gin"
)

type API struct {
	server *httpx.Server

	rootGroup *gin.RouterGroup
}

func New() *API {
	return &API{}
}

func (a API) BindServer(server *httpx.Server, apiPath string) {
	a.server = server

	a.initRoutes(apiPath)
}
