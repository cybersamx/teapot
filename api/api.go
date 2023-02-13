package api

import (
	"github.com/cybersamx/teapot/httpx"
)

type API struct {
	server *httpx.Server
}

func New() *API {
	return &API{}
}

func (a API) BindServer(server *httpx.Server, apiPath string) {
	a.server = server

	a.initRoutes(apiPath)
}
