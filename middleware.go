package tracing

import (
    "time"
    "context"
    "net/http"
    sentryhttp "github.com/getsentry/sentry-go/http"
    "github.com/openrm/module-tracing-golang/opentracing"
    "github.com/openrm/module-tracing-golang/log"
)

const (
    defaultTraceHeader = "Sentry-Trace"
)

type ContextKey struct {
    Namespace string
}

var SpanContextKey = ContextKey{"span"}

var traceHeader = defaultTraceHeader

func SetTraceHeader(v string) {
    traceHeader = v
}

var sentryHandlerOptions = sentryhttp.Options{
    Repanic: true,
    WaitForDelivery: false,
    Timeout: 2 * time.Second,

}

func SetSentryHandlerOptions(opts sentryhttp.Options) {
    sentryHandlerOptions.WaitForDelivery = opts.WaitForDelivery
    sentryHandlerOptions.Timeout = opts.Timeout

}

func Middleware(handler http.Handler) http.Handler {
    sentryHandler := sentryhttp.New(sentryHandlerOptions)

    loggingHandler := log.Handler(func(ctx context.Context) *opentracing.Span {
        if sp, ok := ctx.Value(SpanContextKey).(*opentracing.Span); ok {
            return sp
        }
        return nil
    })

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if traceString := r.Header.Get(traceHeader); traceString != "" {
            span := opentracing.NewFromTraceParent(traceString)
            if span != nil {
                ctx := context.WithValue(r.Context(), SpanContextKey, span)
                r = r.WithContext(ctx)
            }
        }

        handler = sentryHandler.Handle(handler)
        handler = loggingHandler(handler)

        handler.ServeHTTP(w, r)
    })
}
