package errors

import (
    "net/http"
    "github.com/pkg/errors"
)

type CodeError interface {
    StatusCode() int
}

func NewWithStatusCode(message string, code int) error {
    return &withStatusCode{
        errors.New(message),
        code,
    }
}

func WithStatusCode(err error, code int) error {
    if err == nil {
        return nil
    }
    return &withStatusCode{
        err,
        code,
    }
}

func extractStatusCode(err error) int {
    if err, ok := err.(*withStatusCode); ok {
        if err.StatusCode() > 0 {
            return err.StatusCode()
        }
    }

    if err, ok := err.(causer); ok {
        return extractStatusCode(err.Cause())
    }

    return http.StatusInternalServerError
}

func ExtractStatusCode(err error) int {
    return extractStatusCode(err)
}

type withStatusCode struct {
    error
    status int
}

func (w *withStatusCode) Cause() error {
    return w.error
}

func (w *withStatusCode) StatusCode() int {
    return w.status
}
