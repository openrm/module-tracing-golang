package tracing

import (
    "net/http"
    "github.com/getsentry/sentry-go"
)

var ErrorContextKey = ContextKey{"error"}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// implements http.Handler
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := f(w, r); err != nil {
        if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
            hub.CaptureException(error(WithStackTrace(err)))
        }

        if w, ok := w.(*ResponseLogger); ok {
            w.WriteError(err)
        }

        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
