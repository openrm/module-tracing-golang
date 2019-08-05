package tracing

import (
    "regexp"
    "context"
    "net/http"
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

var traceParentPattern = regexp.MustCompile(`^[ \t]*([0-9a-f]{32})?-?([0-9a-f]{16})?-?([01])?[ \t]*$`)

func extractSpan(v string) *Span {
    if matches := traceParentPattern.FindStringSubmatch(v); len(matches) > 2 {
        return &Span{ TraceId: matches[1], SpanId: matches[2] }
    }
    return nil
}

func fromTraceParent(v string) *Span {
    if parent := extractSpan(v); parent != nil {
        span := newSpanFromParent(*parent)
        return &span
    }
    return nil
}

func Middleware(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if traceString := r.Header.Get(traceHeader); traceString != "" {
            span := fromTraceParent(traceString)
            if span != nil {
                ctx := context.WithValue(r.Context(), SpanContextKey, span)
                r = r.WithContext(ctx)
            }
        }
        handler.ServeHTTP(w, r)
    })
}
