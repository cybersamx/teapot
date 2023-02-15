package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cybersamx/teapot/httpx"
	"github.com/cybersamx/teapot/model"
	"github.com/gin-gonic/gin"
)

// Inner auth errors

var (
	ErrAuthHeaderMissing    = errors.New("missing authorization header")
	ErrBearerInvalidMissing = errors.New("invalid or missing bearer")
	ErrCredentialsMissing   = errors.New("missing login credentials")
	ErrCredentialsWrong     = errors.New("wrong login credentials")
	ErrPermissionDenied     = errors.New("insufficient permission")
)

// Internal errors

var (
	ErrMissingClientError = errors.New("missing client error in the context error chain")
)

type ErrorResponse struct {
	Message string   `json:"message,omitempty"`
	Code    int      `json:"code,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// --- Client Error Constructors for the API ---

// CREDIT: https://github.com/kubeflow/pipelines
// The helper functions here are inspired/based by Kubeflow with some modification.
// Licensed under the Apache 2.0 license.

func NewNotFoundErrorf(err error, format string, args ...any) *model.ClientError {
	message := fmt.Sprintf(format, args...)

	return model.NewClientError(
		fmt.Errorf("not found error - %s; %w", message, err),
		http.StatusNotFound,
		message,
	)
}

func NewConflictErrorf(err error, format string, args ...any) *model.ClientError {
	message := fmt.Sprintf(format, args...)

	return model.NewClientError(
		fmt.Errorf("conflict error - %s; %w", message, err),
		http.StatusConflict,
		message,
	)
}

func NewBadRequestErrorf(err error, format string, args ...any) *model.ClientError {
	message := fmt.Sprintf(format, args...)

	return model.NewClientError(
		fmt.Errorf("bad request error - %s; %w", message, err),
		http.StatusBadRequest,
		message,
	)
}

func NewForbiddenError(err error, resource string) *model.ClientError {
	message := fmt.Sprintf("you do not have the permissions to access %v", resource)

	return model.NewClientError(
		fmt.Errorf("forbidden error - %s; %w", message, err),
		http.StatusForbidden,
		message,
	)
}

func NewUnauthorizedErrorf(err error, format string, args ...any) *model.ClientError {
	message := fmt.Sprintf(format, args...)

	return model.NewClientError(
		fmt.Errorf("unauthorized error - %s; %w", message, err),
		http.StatusUnauthorized,
		message,
	)
}

func NewInternalServerErrorf(err error, format string, args ...any) *model.ClientError {
	message := fmt.Sprintf(format, args...)

	// For internal errors, we don't expose any details to the client.
	return model.NewClientError(
		fmt.Errorf("InternalServerError - %s; %w", message, err),
		http.StatusInternalServerError,
		"internal server error",
	)
}

// --- Commonly Used Client Errors ---

// Malformed = not well structured object eg. {"name"}, not a json object.
// Invalid = well-formed object but with the wrong data fields.
// eg. {"name": "sam"}, we expect a product object but got a person object.

func newInvalidMissingBearer() *model.ClientError {
	return NewBadRequestErrorf(ErrBearerInvalidMissing, "invalid or missing bearer")
}

func newBodyError(err error) *model.ClientError {
	return NewBadRequestErrorf(err, "missing or malformed request body data")
}

func newWrongCredentialsError() *model.ClientError {
	return NewUnauthorizedErrorf(ErrCredentialsWrong,
		"can't log user in with the credentials")
}

func newValidationError(err error) *model.ClientError {
	return NewBadRequestErrorf(err, "data validation error")
}

func newPermissionError(u *url.URL) *model.ClientError {
	return NewForbiddenError(ErrPermissionDenied, u.Path)
}

// --- HTTP Handlers ---

func isInternalError(code int) bool {
	return code >= http.StatusInternalServerError && code < 600
}

func pushClientError(ctx *gin.Context, cerr *model.ClientError) {
	gerr := ctx.Error(cerr)

	// Constrain to just private (internal) and public (external) errors.
	if isInternalError(cerr.StatusCode()) {
		gerr.SetType(gin.ErrorTypePrivate)
		return
	}

	gerr.SetType(gin.ErrorTypePublic)
}

func (a API) initErrorHandling() {
	a.server.Logger().Info("Initializing logging middleware")

	a.rootGroup.Use(a.useErrorHandling())
}

func (a API) useErrorHandling() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Run Next() now so that http handlers are executed first.
		ctx.Next()

		// Skip if no errors were recorded.
		if len(ctx.Errors) == 0 {
			return
		}

		var (
			foundCErr           *model.ClientError
			foundErr            error
			internalErrors      []error
			externalErrMessages []string
		)

		// Go through the errors collected in the context chain.
		// 1. Return all external error message to the response.
		// 2. Log all the internal server errors.
		for i := 0; i < len(ctx.Errors); i++ {
			foundErr = ctx.Errors[i].Err

			cerr, ok := foundErr.(*model.ClientError)
			if ok {
				if foundCErr == nil {
					foundCErr = cerr
				}

				// Internal errors.
				if cerr.StatusCode() >= http.StatusInternalServerError && cerr.StatusCode() < 600 {
					internalErrors = append(internalErrors, cerr)
					continue
				}

				// External errors.
				externalErrMessages = append(externalErrMessages, cerr.Message())
				continue
			}

			internalErrors = append(internalErrors, foundErr)
		}

		if foundCErr == nil {
			foundCErr = NewInternalServerErrorf(ErrMissingClientError, "need to have one client error")
		}

		// Log if there's an internal error.
		if len(internalErrors) > 0 {
			obj := httpx.GetContextObject(ctx)
			params := map[string]any{
				"method":      ctx.Request.Method,
				"request-url": ctx.Request.URL.String(),
				"request-id":  obj.RequestID,
				"status":      foundCErr.StatusCode,
				"errors":      internalErrors,
			}

			a.server.Logger().WithFields(params).Error("API internal (only) errors found")
		}

		// Return error response to client.
		er := ErrorResponse{
			Message: foundCErr.Message(),
			Code:    foundCErr.StatusCode(),
			Errors:  externalErrMessages,
		}

		ctx.JSON(foundCErr.StatusCode(), er)
	}
}
