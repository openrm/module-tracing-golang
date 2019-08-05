package tracing

import (
    "net/http"
    "github.com/getsentry/sentry-go"
)

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// implements http.Handler
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := f(w, r); err != nil {
        if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
            hub.CaptureException(err)
        }

        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
