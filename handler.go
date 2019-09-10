package tracing

import (
    "net/http"
    "encoding/json"
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

        var errBody interface{} = map[string]interface{}{
            "message": err.Error(),
        }

        if err := errors.ExtractResponseError(err); err != nil {
            errBody = err.Body()
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.WriteHeader(errors.ExtractStatusCode(err))

        json.NewEncoder(w).Encode(errBody)
    }
}
