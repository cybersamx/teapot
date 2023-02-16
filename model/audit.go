package model

import (
	"github.com/cybersamx/teapot/common"
)

type Audit struct {
	RequestID     string `json:"request_id,omitempty" db:"request_id"`
	CreatedAt     int64  `json:"created_at" db:"created_at"`
	ClientAgent   string `json:"client_agent,omitempty" db:"client_agent"`
	ClientAddress string `json:"client_address,omitempty" db:"client_address"`
	StatusCode    int    `json:"status_code" db:"status_code"`
	Error         string `json:"error" db:"error"`
	Event         string `json:"event,omitempty" db:"event"`
}

func (a *Audit) PreSave() {
	if a.CreatedAt == 0 {
		a.CreatedAt = common.NowInMilli()
	}
}
