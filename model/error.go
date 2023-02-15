package model

import (
	"errors"
	"fmt"
)

// CREDIT: https://github.com/kubeflow/pipelines
// The error programming model here is inspired/based by Kubeflow with some modification.
// Licensed under the Apache 2.0 license.

// --- ClientError ---

var _ error = (*ClientError)(nil)

// ClientError represents the error with message and status code for an external client.
// Status code can be http status code or grpc code. The wrapped innerErr represents
// internal errors and messages.
type ClientError struct {
	innerErr   error
	message    string
	statusCode int
}

func NewClientError(innerErr error, statusCode int, message string) *ClientError {
	return &ClientError{
		innerErr:   innerErr,
		statusCode: statusCode,
		message:    message,
	}
}

func NewClientErrorf(innerErr error, statusCode int, format string, args ...any) *ClientError {
	message := fmt.Sprintf(format, args...)

	return &ClientError{
		innerErr:   innerErr,
		statusCode: statusCode,
		message:    message,
	}
}

func (ce *ClientError) Error() string {
	return ce.innerErr.Error()
}

func (ce *ClientError) Message() string {
	return ce.message
}

func (ce *ClientError) StatusCode() int {
	return ce.statusCode
}

func (ce *ClientError) Wrap(message string) *ClientError {
	ce.message = message
	ce.innerErr = fmt.Errorf("%s: %w", ce.message, ce.innerErr)

	return ce
}

func (ce *ClientError) Wrapf(format string, args ...any) *ClientError {
	ce.message = fmt.Sprintf(format, args...)
	ce.innerErr = fmt.Errorf("%s: %w", ce.message, ce.innerErr)

	return ce
}

func (ce *ClientError) Is(targetErr error) bool {
	return errors.Is(ce.innerErr, targetErr)
}
