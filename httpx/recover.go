package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (s *Server) initPanicRecovery() {
	s.logger.Info("Initializing panic recovery middleware")

	s.router.Use(s.usePanicRecovery())
}

func (s *Server) usePanicRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, exception any) {
		const msg = "server panic"

		fields := logrus.Fields{
			"method":  ctx.Request.Method,
			"path":    fullPath(ctx.Request),
			"handler": ctx.HandlerName(),
			"agent":   ctx.Request.UserAgent(),
		}

		if errMsg, ok := exception.(string); ok {
			fields["err-msg"] = errMsg
			s.logger.WithFields(fields).Errorln(msg)
			ctx.String(http.StatusInternalServerError, errMsg)
			ctx.Abort()
			return
		}

		if err, ok := exception.(error); ok {
			s.logger.WithError(err).WithFields(fields).Errorln(msg)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		s.logger.WithFields(fields).Errorln(msg)
		ctx.AbortWithStatus(http.StatusInternalServerError)
	})
}
