package api

import (
	"errors"
	"testing"

	"github.com/cybersamx/teapot/model"
	"github.com/stretchr/testify/assert"
)

func TestError_NewClientError(t *testing.T) {
	errInner := errors.New("inner error")

	tests := []struct {
		description string
		cerr        *model.ClientError
		statusCode  int
		wantMessage string
	}{
		{
			"NewBadRequestErrorf",
			NewBadRequestErrorf(errInner, "invalid pod-name pod-123"),
			400,
			"invalid pod-name pod-123",
		},
		{
			"NewNotFoundError",
			NewNotFoundErrorf(errInner, "can't find %s", "pod-123"),
			404,
			"can't find pod-123",
		},
		{
			"NewConflictError",
			NewConflictErrorf(errInner, "resource %s already exists", "network-abc"),
			409,
			"resource network-abc already exists",
		},
		{
			"NewConflictError",
			NewConflictErrorf(errInner, "resources %s %s already exist", "network-abc", "pod-123"),
			409,
			"resources network-abc pod-123 already exist",
		},
		{
			"NewForbiddenError",
			NewForbiddenError(errInner, "network-abc"),
			403,
			"you do not have the permissions to access network-abc",
		},
		{
			"NewUnauthorizedErrorf",
			NewUnauthorizedErrorf(errInner, "user %s cannot be authenticated", "spatel"),
			401,
			"user spatel cannot be authenticated",
		},
		{
			"NewInternalServerErrorf",
			NewInternalServerErrorf(errInner, "failed to connect to database %s", "sqlite"),
			500,
			"internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert.Error(t, tc.cerr)
			assert.Equal(t, tc.statusCode, tc.cerr.StatusCode())
			assert.Equal(t, tc.wantMessage, tc.cerr.Message())
			assert.True(t, errors.Is(tc.cerr, errInner))
		})
	}
}
