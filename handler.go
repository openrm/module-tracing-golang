package tracing

import (
    "net/http"
    "github.com/getsentry/sentry-go"
    "github.com/openrm/module-tracing-golang/errors"
    "github.com/openrm/module-tracing-golang/log"
)

var ErrorContextKey = ContextKey{"error"}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// implements http.Handler
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := f(w, r); err != nil {
        if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
            hub.CaptureException(error(errors.WithStackTrace(err)))
        }

        if w, ok := w.(*log.ResponseLogger); ok {
            w.WriteError(err)
        }

        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
