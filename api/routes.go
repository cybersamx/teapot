package api

func (a API) initMiddleware() {
	a.initErrorHandling()
}

func (a API) initRoutes(apiPath string) {
	a.server.Logger().
		WithFields(map[string]any{"site-url": a.server.Config().HTTP.SiteURL}).
		Info("Initializing server routes")

	a.rootGroup = a.server.Router().Group(apiPath)

	a.initMiddleware()

	a.rootGroup.GET("/ping", a.handlePing())
	a.rootGroup.GET("/err", a.handleErr())
	a.rootGroup.GET("/panic", a.handlePanic())
}
