package log

import (
    "time"
    "context"
    "net/http"
    log "github.com/sirupsen/logrus"
    "github.com/openrm/module-tracing-golang/opentracing"
)

type contextKey struct {}

var (
    LoggerContextKey = contextKey{}
)

const (
    messageRequestHandled = "tracing: request handled"
    messageCaughtPanic = "tracing: caught panic"
    errorKey = "err"
)

func Handler(extractSpan func(context.Context) *opentracing.Span) func(http.Handler) http.Handler {
    return func(handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, re := range excludePatterns {
                if re.MatchString(r.URL.Path) {
                    handler.ServeHTTP(w, r)
                    return
                }
            }

            start := time.Now()

            entry := logger.WithFields(log.Fields{
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

            if sp := extractSpan(r.Context()); sp != nil {
                span := *sp
                entry = entry.WithFields(log.Fields{
                    "span": span.JSON(true),
                    "traceId": span.TraceId,
                    "spanId": span.SpanId,
                })
            }

            ctx := r.Context()
            ctx = context.WithValue(ctx, LoggerContextKey, NewLogger(entry))

            r = r.WithContext(ctx)
            l := NewResponseLogger(w)

            defer func() {
                if err := recover(); err != nil {
                    entry.WithField(errorKey, parsePanic(err)).Error(messageCaughtPanic)
                }
            }()

            handler.ServeHTTP(l, r)

            if err := l.Error; err != nil {
                entry = entry.WithField(errorKey, parseError(err))
            }

            if r.Form != nil {
                entry = entry.WithField("form", r.Form)
            }

            if r.MultipartForm != nil {
                entry = entry.WithField("files", parseMultipartForm(r.MultipartForm))
            }

            entry = entry.WithFields(log.Fields{
                "responseTime": float64(time.Since(start).Nanoseconds()) / 1e6, // ms
                "status": l.Status,
                "responseHeaders": filterHeader(l.Header()),
                "responseContentLength": l.Size,
            })

            entry.Info(messageRequestHandled)
        })
    }
}

