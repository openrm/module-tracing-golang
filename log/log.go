package log

import (
    "time"
    "context"
    "net/http"
    log "github.com/sirupsen/logrus"
    "go.opencensus.io/trace"
)

type contextKey struct {}
type extraKey struct {
    name string
}

var (
    LoggerContextKey = contextKey{}
    EventIdKey = extraKey{"eventId"}
)

const (
    messageRequestHandled = "tracing: request handled"
    messageCaughtPanic = "tracing: caught panic"
    errorKey = "err"
)

func toMap(span *trace.Span) map[string]interface{} {
    if span == nil {
        return nil
    }
    sc := span.SpanContext()
    return map[string]interface{}{
        "traceId": sc.TraceID.String(),
        "spanId": sc.SpanID.String(),
        "sampled": sc.IsSampled(),
    }
}

func Handler() func(http.Handler) http.Handler {
    return func(handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, re := range excludePatterns {
                if re.MatchString(r.URL.Path) {
                    handler.ServeHTTP(w, r)
                    return
                }
            }

            start := time.Now()

            var entry *log.Entry = globalLogger.WithFields(nil)

            if span := trace.FromContext(r.Context()); span != nil {
                entry = entry.WithField("span", toMap(span))
            }

            ctx := r.Context()
            ctx = context.WithValue(ctx, LoggerContextKey, NewLogger(entry))

            entry = entry.WithFields(log.Fields{
                "method": r.Method,
                "protocol": r.Proto,
                "url": r.RequestURI,
                "headers": filterHeader(r.Header),
                "remoteAddress": r.RemoteAddr,
                "hostname": r.Host,
                "referer": r.Referer(),
                "userAgent": r.UserAgent(),
                "cookies": parseCookies(r.Cookies()),
                "contentLength": r.ContentLength,
            })

            r = r.WithContext(ctx)
            l := NewResponseLogger(w)

            defer func() {
                if err := recover(); err != nil {
                    entry.WithField(errorKey, parsePanic(err)).Error(messageCaughtPanic)
                }
            }()

            handler.ServeHTTP(l, r)

            if l.err != nil {
                entry = entry.WithField(errorKey, parseError(l.err))
            }

            if eventId, ok := l.getExtra(EventIdKey).(string); ok {
                entry = entry.WithField("sentry", map[string]string{
                    "eventId": eventId,
                })
            }

            if r.Form != nil {
                entry = entry.WithField("form", r.Form)
            }

            if r.MultipartForm != nil {
                entry = entry.WithField("files", parseMultipartForm(r.MultipartForm))
            }

            entry = entry.WithFields(log.Fields{
                "responseTime": float64(time.Since(start).Nanoseconds()) / 1e6, // ms
                "status": l.status,
                "responseHeaders": filterHeader(l.Header()),
                "responseContentLength": l.size,
            })

            entry.Info(messageRequestHandled)
        })
    }
}

