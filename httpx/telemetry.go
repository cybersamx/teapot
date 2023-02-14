package httpx

import (
	"fmt"
	"net/http"
	"os"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sirupsen/logrus"
)

const (
	storeCheckInitialDelay    = 2 * time.Second
	storeCheckExecutionPeriod = 15 * time.Second
)

func fullPath(req *http.Request) string {
	if req == nil {
		return ""
	}

	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, req.URL.RawQuery)
	}

	return path
}

func (s *Server) initPrometheus() {
	s.logger.Info("Initializing prometheus middleware")

	if s.cfg.HTTP.EnablePrometheus {
		s.promRegistry = prometheus.NewRegistry()
	}

	// Set up collectors.
	goCollector := collectors.NewGoCollector()

	processCollector := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		PidFn: func() (int, error) {
			return os.Getpid(), nil
		},
	})

	counterCollector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total http request count",
	}, []string{"method", "path", "status"})

	s.promRegistry.MustRegister(goCollector, processCollector, counterCollector)
	s.logger.WithFields(map[string]any{
		"collectors": []string{"go", "process", "http_requests_total"},
	}).Info("Registered prometheus")

	//s.router.Use(s.usePrometheusCounter(counterCollector))
}

//func (s *Server) usePrometheusCounter(collector *prometheus.CounterVec) gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		ctx.Next()
//
//		collector.With(prometheus.Labels{
//			"method": ctx.Request.Method,
//			"path":   ctx.Request.URL.Path,
//			"status": strconv.Itoa(ctx.Writer.Status()),
//		}).Inc()
//	}
//}

func (s *Server) initHealthCheck() {
	s.healthcheck = gosundheit.New()

	storeCheck, err := checks.NewPingCheck("datastore", s.Store())
	if err == nil {
		err = s.healthcheck.RegisterCheck(
			storeCheck,
			gosundheit.InitialDelay(storeCheckInitialDelay),
			gosundheit.ExecutionPeriod(storeCheckExecutionPeriod),
		)
	}
}

//func (s *Server) initLogging() {
//	s.logger.Info("Initializing logging middleware")
//
//	s.router.Use(s.useLogging())
//}

//func (s *Server) useLogging() gin.HandlerFunc {
//	// Custom logger for all incoming requests based on gin-gonic standard log handler.
//	// https://github.com/gin-gonic/gin/blob/master/logger.go
//	return func(ctx *gin.Context) {
//		start := time.Now()
//
//		ctx.Next()
//
//		// For now, log the requests to all paths.
//		latency := time.Now().Sub(start)
//		status := ctx.Writer.Status()
//
//		path := fullPath(ctx.Request)
//
//		fields := logrus.Fields{
//			"latency":   latency,
//			"client-ip": ctx.ClientIP(),
//			"agent":     ctx.Request.UserAgent(),
//			"body":      ctx.Writer.Size(),
//		}
//
//		errs := ctx.Errors.ByType(gin.ErrorTypePrivate)
//		if len(errs) > 0 {
//			fields["error"] = errs.String()
//		}
//
//		msg := ""
//		if s.cfg.LogFormat == "text" {
//			msg = fmt.Sprintf("%3d - %-7s %q", status, ctx.Request.Method, path)
//		} else {
//			fields["status"] = status
//			fields["method"] = ctx.Request.Method
//			fields["path"] = path
//
//			msg = fmt.Sprintf("%d - %s %s", status, ctx.Request.Method, path)
//		}
//
//		entry := s.logger.WithFields(fields)
//
//		if len(errs) > 0 {
//			entry.Errorln(msg)
//			return
//		}
//
//		if status < http.StatusOK || status >= http.StatusMultipleChoices {
//			entry.Warnln(msg)
//			return
//		}
//
//		entry.Info(msg)
//	}
//}

func (s *Server) handleLogging(ctx *Context, w *responseWriterWrapper, r *http.Request) {
	// For now, log the requests to all paths.
	latency := time.Now().Sub(ctx.startTime)

	path := fullPath(r)

	fields := logrus.Fields{
		"latency":   latency,
		"client-ip": ctx.clientIP,
		"agent":     r.UserAgent(),
		"body":      w.Size(),
	}

	s.logger.WithFields(fields).Infof("%d - %s %s", w.Status(), r.Method, path)
}

//func (s *Server) initTracing() {
//	s.logger.Info("Initializing tracing middleware")
//
//	s.router.Use(s.useTracing())
//}

//func (s *Server) useTracing() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		obj := GetContextObject(ctx)
//		obj.RequestID = uuid.New().String()
//		SetContextObject(ctx, obj)
//
//		ctx.Writer.Header().Add(HeaderXRequestID, obj.RequestID)
//
//		ctx.Next()
//	}
//}

func (s *Server) handleTracing(handlerFn http.Handler) {
	// Setup

}

func (s *Server) handleHealthCheck() HandlerFunc {
	return func(ctx *Context, w *responseWriterWrapper, r *http.Request) {
		renderJSON(w, http.StatusOK, map[string]any{"status": "OK"})
	}
}
