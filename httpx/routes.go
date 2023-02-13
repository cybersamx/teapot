package httpx

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	ghttp "github.com/AppsFlyer/go-sundheit/http"
)

func (s *Server) Handle(path string, handler http.Handler) {
	s.router.Handle(path, handler)
}

func (s *Server) HandlerFunc(path string, handlerFn http.HandlerFunc) {
	s.Handle(path, handlerFn)
}

func (s *Server) initMiddleware() {
	// Add all our middleware for the server layer here.
	//s.initPanicRecovery()
	//s.initLogging()
	s.initHealthCheck()
	//s.initTracing()
	//s.initCORS()

	if s.cfg.HTTP.EnablePrometheus {
		s.initPrometheus()
	}
}

func (s *Server) initRoutes() {
	s.logger.WithFields(map[string]any{"site-url": s.cfg.HTTP.SiteURL}).Info("Initializing server routes")

	// Return 405 if method is not supported. By default, gin returns 404.
	//s.router.HandleMethodNotAllowed = true

	// Middleware.
	s.initMiddleware()

	// Root routes.
	s.router.Handle("/health", APIHandler(s.handleHealthCheck())).Methods(http.MethodGet)
	s.router.Handle("/health/live", APIHandler(s.handleHealthCheck())).Methods(http.MethodGet)
	s.router.Handle("/health/ready", APIHandler(WrapHandlerFunc(ghttp.HandleHealthJSON(s.healthcheck)))).Methods(http.MethodGet)
	//s.router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.promRegistry, promhttp.HandlerOpts{})))

	// Profiler routes.
	if s.Config().Mode != "production" && s.Config().HTTP.EnableProfiler {
		s.Logger().Info("Initializing the profilerGroup")

		pprofRoutes := s.router.PathPrefix("/pprof").Subrouter()
		pprofRoutes.Handle("/vars", expvar.Handler()).Methods(http.MethodGet)
		pprofRoutes.HandleFunc("", pprof.Index)

		//profilerGroup.GET("/", gin.WrapF(pprof.Index))
		//profilerGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		//profilerGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		//profilerGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		//profilerGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		//profilerGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		//profilerGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		//profilerGroup.GET("/profile", gin.WrapF(pprof.Profile))
		//profilerGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		//profilerGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
		//profilerGroup.GET("/trace", gin.WrapF(pprof.Trace))
	}
}
