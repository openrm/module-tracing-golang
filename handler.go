package tracing

import (
    "net/http"
    "encoding/json"
    "go.opencensus.io/trace"
    "github.com/getsentry/sentry-go"
    "github.com/openrm/module-tracing-golang/errors"
    "github.com/openrm/module-tracing-golang/log"
    "github.com/openrm/module-tracing-golang/propagation"
)

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

const (
    sentrySpanKey = "span"
    sentryParentSpanContextKey = "parent_span_context"
)

// implements http.Handler
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := f(w, r); err != nil {
        var eventId *sentry.EventID

        if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
            scope := hub.Scope()
            scope.SetExtra(sentrySpanKey, trace.FromContext(r.Context()))

            format := propagation.HTTPFormat{ Header: traceHeader }
            if sc, ok := format.SpanContextFromRequest(r); ok {
                scope.SetExtra(sentryParentSpanContextKey, sc)
            }

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

        if rerr := errors.ExtractResponseError(err); rerr != nil {
            if rerr.StatusCode() < http.StatusInternalServerError {
                err = errors.WithStatusCode(err, rerr.StatusCode())
            }
            errBody = rerr.Body()
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.WriteHeader(errors.ExtractStatusCode(err))

        json.NewEncoder(w).Encode(errBody)
    }
}
