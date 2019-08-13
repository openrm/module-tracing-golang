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
)

type stackTracer interface {
    error
    StackTrace() errors.StackTrace
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

            entry := logger
            start := time.Now()

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

            if sp, ok := r.Context().Value(tracing.SpanContextKey).(*tracing.Span); ok {
                span := *sp
                entry = entry.WithFields(log.Fields{
                    "span": span.JSON(),
                    "traceId": span.TraceId,
                    "spanId": span.SpanId,
                })
            }

            ctx := context.WithValue(r.Context(), LoggerContextKey, entry)

            l := tracing.NewResponseLogger(w)
            r = r.WithContext(ctx)

            handler.ServeHTTP(l, r)

            if err := l.Error; err != nil {
                errMap := map[string]interface{}{
                    "message": err.Error(),
                }

                if err, ok := err.(stackTracer); ok {
                    st := err.StackTrace()
                    stackTrace := make([]string, len(st))

                    for i, frame := range st {
                        stackTrace[i] = fmt.Sprintf("%+v", frame)
                    }

                    errMap["stack"] = stackTrace
                }

                entry = entry.WithField("err", errMap)
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

