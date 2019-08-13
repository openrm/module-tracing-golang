package log

import (
    "fmt"
    "time"
    "context"
    "net/http"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/openrm/module-tracing-golang"
)

var (
    LoggerContextKey = tracing.ContextKey{"logger"}
)

const (
    messageRequestHandled = "tracing: request handled"
    errorKey = "err"
)


type stackTracer interface {
    error
    StackTrace() errors.StackTrace
    Cause() error
}

func traceError(err error, seen []string) []string {
    if err, ok := err.(stackTracer); ok {
        st := err.StackTrace()

        var stackTrace []string
        unique := len(seen) == 0

        for i := 0; i < len(st); i++ {
            frame := st[len(st) - 1 - i]
            str := fmt.Sprintf("%+v", frame)
            if unique || len(seen) > i && seen[len(seen) - 1 - i] != str {
                unique = true
                stackTrace = append([]string{str}, stackTrace...)
            }
        }

        if err.Cause() == nil {
            return stackTrace
        }

        return append(traceError(err.Cause(), append(stackTrace, seen...)), stackTrace...)
    }

    return []string{}
}

func parseError(err error) map[string]interface{} {
    errMap := map[string]interface{}{
        "message": err.Error(),
    }

    if err, ok := err.(stackTracer); ok {
        errMap["stack"] = traceError(err, []string{})
    }

    return errMap
}

func Handler(logger log.FieldLogger) mux.MiddlewareFunc {
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

            if sp, ok := r.Context().Value(tracing.SpanContextKey).(*tracing.Span); ok {
                span := *sp
                entry = entry.WithFields(log.Fields{
                    "span": span.JSON(),
                    "traceId": span.TraceId,
                    "spanId": span.SpanId,
                })
            }

            ctx := r.Context()
            ctx = context.WithValue(ctx, LoggerContextKey, NewLogger(entry))

            r = r.WithContext(ctx)
            l := tracing.NewResponseLogger(w)

            handler.ServeHTTP(l, r)

            if err := l.Error; err != nil {
                errMap := parseError(err)
                entry = entry.WithField(errorKey, errMap)
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

