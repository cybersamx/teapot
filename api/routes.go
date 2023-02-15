package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (a API) initMiddleware() {
	a.initErrorHandling()
}

func (a API) initRoutes(apiPath string) {
	a.server.Logger().
		WithFields(map[string]any{"site-url": a.server.Config().HTTP.SiteURL}).
		Info("Initializing server routes")

	a.rootGroup = a.server.Router().Group(apiPath)

	a.initMiddleware()

	// Simulate a happy path.
	a.rootGroup.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	// Simulate an error path.
	a.rootGroup.GET("/err", func(ctx *gin.Context) {
		// Only non-5xx errors will be returned to the client while 5xx errors will be logged.
		pushClientError(ctx, NewConflictErrorf(errors.New("root error"), "can't create user with %v", "my-username"))
		pushClientError(ctx, NewNotFoundErrorf(errors.New("root error"), "can't find user with %v", "my-username"))
		pushClientError(ctx, NewInternalServerErrorf(errors.New("internal error"), "can't open db"))
		ctx.Abort()
	})

	// Simulate a panic.
	a.rootGroup.GET("/panic", func(ctx *gin.Context) {
		logrus.Panic("panic")
	})
}
