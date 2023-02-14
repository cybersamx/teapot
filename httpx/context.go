package httpx

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Context struct {
	logger    *logrus.Logger
	requestID string
	clientIP  string
	path      string
	userAgent string
	bodySize  int
	startTime time.Time
}
