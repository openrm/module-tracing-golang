package tracing

import (
    "time"
    "net/http"
    sentryhttp "github.com/getsentry/sentry-go/http"
    "github.com/openrm/module-tracing-golang/log"
    "github.com/openrm/module-tracing-golang/propagation"
    "go.opencensus.io/plugin/ochttp"
    "go.opencensus.io/trace"
)

type ContextKey struct {
    Namespace string
}

var SpanContextKey = ContextKey{"span"}

var sentryHandlerOptions = sentryhttp.Options{
    Repanic: true,
    WaitForDelivery: false,
    Timeout: 2 * time.Second,

}

func SetSentryHandlerOptions(opts sentryhttp.Options) {
    sentryHandlerOptions.WaitForDelivery = opts.WaitForDelivery
    sentryHandlerOptions.Timeout = opts.Timeout

}

func Middleware() func(http.Handler) http.Handler {
    sentryHandler := sentryhttp.New(sentryHandlerOptions)
    loggingHandler := log.Handler(log.Options{ TraceHeader: traceHeader })

    return func(handler http.Handler) http.Handler {
        handler = sentryHandler.Handle(handler)
        handler = loggingHandler(handler)
        return &ochttp.Handler{
            Propagation: &propagation.HTTPFormat{
                Header: traceHeader,
            },
            Handler: handler,
            GetStartOptions: func(r *http.Request) trace.StartOptions {
                for _, re := range log.GetExcludePatterns() {
                    if re.MatchString(r.URL.Path) {
                        return trace.StartOptions{ Sampler: trace.NeverSample() }
                    }
                }
                return trace.StartOptions{}
            },
        }
    }
}
