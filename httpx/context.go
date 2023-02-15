package httpx

import (
	"github.com/gin-gonic/gin"
)

const (
	ctxObjectKey = "request-object"
)

type ContextObject struct {
	RequestID string
}

func GetContextObject(ctx *gin.Context) ContextObject {
	val, ok := ctx.Get(ctxObjectKey)
	if !ok {
		val = ContextObject{}
	}

	obj, ok := val.(ContextObject)
	if !ok {
		obj = ContextObject{}
	}

	return obj
}

func SetContextObject(ctx *gin.Context, obj ContextObject) {
	ctx.Set(ctxObjectKey, obj)
}
