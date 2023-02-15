package httpx

import (
	"time"

	"github.com/gin-contrib/cors"
)

func (s *Server) initCORS() {
	// * (wildcard) = allow any domains
	// "" (empty) = disable cors
	// domain names = allow these domains

	allowAll := false
	for _, origin := range s.cfg.HTTP.AllowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
	}

	if len(s.cfg.HTTP.AllowedOrigins) == 0 {
		return
	}

	s.logger.Info("Initializing cors middleware")

	// Using cors handler from gin.
	ccfg := cors.Config{
		AllowAllOrigins:  allowAll,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	if !allowAll {
		ccfg.AllowOrigins = s.cfg.HTTP.AllowedOrigins
	}

	s.router.Use(cors.New(ccfg))
}
