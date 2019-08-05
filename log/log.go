package log

import (
    "time"
    "context"
    "regexp"
    "net/http"
    log "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/openrm/module-tracing-golang"
)

var (
    LoggerContextKey = tracing.ContextKey{"logger"}
    excludePatterns = []*regexp.Regexp{}
)

func SetExcludePatterns(exprs ...string) {
    for _, expr := range exprs {
        re, err := regexp.Compile(expr)
        if err == nil {
            excludePatterns = append(excludePatterns, re)
        }
    }
}

const (
    messageRequestHandled = "tracing: request handled"
)

func Handler(logger *log.Logger) mux.MiddlewareFunc {
    return func(handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, re := range excludePatterns {
                if re.MatchString(r.URL.Path) {
                    handler.ServeHTTP(w, r)
                    return
                }
            }

            entry := log.NewEntry(logger)
            start := time.Now()

            entry = entry.WithField("headers", r.Header)

            entry = entry.WithFields(log.Fields{
                "method": r.Method,
                "protocol": r.Proto,
                "uri": r.RequestURI,
            })

            if sp, ok := r.Context().Value(tracing.SpanContextKey).(*tracing.Span); ok {
                span := *sp
                entry = entry.WithFields(log.Fields{
                    "traceId": span.TraceId,
                    "spanId": span.SpanId,
                })
            }

            ctx := context.WithValue(r.Context(), LoggerContextKey, entry)

            l := newResponseLogger(w)
            handler.ServeHTTP(l, r.WithContext(ctx))

            entry = entry.WithFields(log.Fields{
                "responseTime": time.Since(start),
                "status": l.status,
                "contentLength": l.size,
            })

            entry.Info(messageRequestHandled)
        })
    }
}

