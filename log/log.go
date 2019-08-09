package log

import (
    "fmt"
    "time"
    "context"
    "regexp"
    "net/http"
    "mime/multipart"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/openrm/module-tracing-golang"
)

const Mask = "[Filtered]"

var (
    LoggerContextKey = tracing.ContextKey{"logger"}
    excludePatterns = []*regexp.Regexp{}
    filterHeaderPatterns = []*regexp.Regexp{
        regexp.MustCompile(`[Aa]uthorization`),
    }
)

func SetExcludePatterns(exprs ...string) {
    for _, expr := range exprs {
        re, err := regexp.Compile(expr)
        if err == nil {
            excludePatterns = append(excludePatterns, re)
        }
    }
}

func SetFilterHeaderPatterns(exprs ...string) {
    for _, expr := range exprs {
        re, err := regexp.Compile(expr)
        if err == nil {
            filterHeaderPatterns = append(excludePatterns, re)
        }
    }
}

const (
    messageRequestHandled = "tracing: request handled"
)

type stackTracer interface {
    error
    StackTrace() errors.StackTrace
}

func parseMultipartForm(form *multipart.Form) map[string]interface{} {
    parsedForm := make(map[string]interface{}, len(form.File))

    // the content of form.Value is already set in http.Request.Form
    for k, vs := range form.File {
        files := make([]map[string]interface{}, len(vs))
        for i, f := range vs {
            files[i] = map[string]interface{}{
                "filename": f.Filename,
                "headers": f.Header,
                "size": f.Size,
            }
        }
        parsedForm[k] = files
    }

    return parsedForm
}

func parseCookies(cookies []*http.Cookie) []string {
    serialized := make([]string, len(cookies))
    for i, cookie := range cookies {
        serialized[i] = cookie.String()
    }
    return serialized
}

func filterHeader(headers http.Header) map[string]string {
    filtered := make(map[string]string)
    for k, _ := range headers {
        var value string = headers.Get(k)

        for _, re := range filterHeaderPatterns {
            if re.MatchString(k) {
                value = Mask
            }
        }

        filtered[k] = value
    }
    return filtered
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

            entry = entry.WithField("headers", filterHeader(r.Header))

            entry = entry.WithFields(log.Fields{
                "method": r.Method,
                "protocol": r.Proto,
                "url": r.RequestURI,
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

                    errMap["stacktrace"] = stackTrace
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
                "responseHeaders": l.Header(),
                "responseContentLength": l.Size,
            })

            entry.Info(messageRequestHandled)
        })
    }
}

