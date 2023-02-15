package httpx

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/AppsFlyer/go-sundheit"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Server struct {
	cfg *model.Config

	datastore    store.Store
	httpd        *http.Server
	router       *gin.Engine
	logger       *logrus.Logger
	promRegistry *prometheus.Registry
	healthcheck  gosundheit.Health

	wg       sync.WaitGroup
	cancel   context.CancelFunc
	doneChan <-chan struct{}
}

func New(datastore store.Store, logger *logrus.Logger, cfg *model.Config) *Server {
	s := &Server{
		logger:    logger,
		datastore: datastore,
	}

	// Initialize gin. Note: gin.SetMode() must be run before gin.New().
	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Set up http server.
	router := gin.New()

	s.cfg = cfg
	s.httpd = &http.Server{
		Addr:    cfg.HTTP.Address,
		Handler: router,
	}
	s.router = router

	logWriter := io.MultiWriter(s.logger.Writer())
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter

	s.initRoutes()

	return s
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Config() *model.Config {
	return s.cfg
}

func (s *Server) Logger() *logrus.Logger {
	return s.logger
}

func (s *Server) Store() store.Store {
	return s.datastore
}

func (s *Server) HTTPServer() *http.Server {
	return s.httpd
}

func (s *Server) Start(ctx context.Context) {
	s.logger.WithFields(map[string]any{"addr": s.cfg.HTTP.Address}).Info("Starting http server")

	// Set up the context.
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()
	s.doneChan = ctx.Done()

	// Start the http server.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		listen, err := net.Listen("tcp", s.cfg.HTTP.Address)
		if err != nil {
			s.logger.WithError(err).Fatalf("failed to initialize tcp listener at address %s", s.cfg.HTTP.Address)
			return
		}

		err = s.httpd.Serve(listen)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.WithError(err).Fatalf("failed to start http server at address %s", s.cfg.HTTP.Address)
			return
		}
	}()
}

func (s *Server) Close(ctx context.Context) {
	s.logger.Infoln("Shutting down http server")

	defer func() {
		if s.cancel == nil {
			return
		}

		s.cancel()
		s.wg.Wait()
		s.cancel = nil
	}()

	// In some tests, we don't start the server and instead rely on ServerHTTP. So the
	// httpd field may be nil.
	if s.httpd == nil {
		return
	}

	if err := s.httpd.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Fatalln("Can't shut down http service")
	}
}

func (s *Server) initRequestID() {
	s.logger.Info("Initializing tracing middleware")

	s.router.Use(s.useRequestID())
}

func (s *Server) useRequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		obj := GetContextObject(ctx)
		obj.RequestID = uuid.New().String()
		SetContextObject(ctx, obj)

		ctx.Writer.Header().Add(HeaderXRequestID, obj.RequestID)

		ctx.Next()
	}
}
