package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_ClientError(t *testing.T) {
	err := errors.New("root error")
	tests := []struct {
		cerr        *ClientError
		wantCode    int
		wantMessage string
	}{
		{
			NewClientError(err, 500, "internal error"),
			500,
			"internal error",
		},
		{
			NewClientErrorf(err, 404, "not found: %s", "text-file"),
			404,
			"not found: text-file",
		},
	}

	for _, tc := range tests {
		assert.NotNil(t, tc.cerr)
		assert.Equal(t, tc.wantMessage, tc.cerr.Message())
		assert.Equal(t, err.Error(), tc.cerr.Error())
		assert.Equal(t, tc.wantCode, tc.cerr.StatusCode())
		assert.True(t, errors.Is(tc.cerr, err))
	}
}
