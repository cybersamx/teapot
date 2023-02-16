package api

import (
	"errors"
	"net/http"

	"github.com/cybersamx/teapot/httpx"
	"github.com/cybersamx/teapot/model"
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

// Since this project is a template for creating other projects, only 3 simple handlers
// are implemented. Replace these handlers with yours.

// handlePing simulates a happy path and capture audits.
func (a API) handlePing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		audit := model.Audit{
			RequestID:     httpx.GetContextObject(ctx).RequestID,
			ClientAgent:   ctx.Request.Header.Get("User-Agent"),
			ClientAddress: ctx.ClientIP(),
			StatusCode:    http.StatusOK,
			Event:         "handlePing",
		}

		_, err := a.server.Store().Audits().Insert(ctx, &audit)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.String(http.StatusOK, "pong")
	}
}

// handleErr simulates an error path.
func (a API) handleErr() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Only non-5xx errors will be returned to the client while 5xx errors will be logged.
		pushClientError(ctx, NewConflictErrorf(errors.New("root error"), "can't create user with %v", "my-username"))
		pushClientError(ctx, NewInternalServerErrorf(errors.New("internal error"), "can't open db"))
		ctx.Abort()
	}
}

// handlePanic simulates a panic.
func (a API) handlePanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		panic("panic")
	}
}
