package tracing

import (
    "net/http"
    "github.com/getsentry/sentry-go"
    "github.com/openrm/module-tracing-golang/errors"
    "github.com/openrm/module-tracing-golang/log"
)

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// implements http.Handler
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := f(w, r); err != nil {
        var eventId *sentry.EventID

        if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
            eventId = hub.CaptureException(error(errors.WithStackTrace(err)))
        }

        if w, ok := w.(*log.ResponseLogger); ok {
            w.WriteError(err)

            if eventId != nil {
                w.WriteExtra(log.EventIdKey, string(*eventId))
            }
        }

        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
