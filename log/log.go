package log

import (
    "time"
    "bytes"
    "context"
    "net/http"
    "encoding/json"
    log "github.com/sirupsen/logrus"
    apachelog "github.com/lestrrat-go/apache-logformat"
    sdlog "cloud.google.com/go/logging"
    "go.opencensus.io/trace"
    "github.com/openrm/module-tracing-golang/propagation"
)

func init() {
}

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

type Options struct {
    TraceHeader string
}

func toMap(sc trace.SpanContext) map[string]interface{} {
    return map[string]interface{}{
        "traceId": sc.TraceID.String(),
        "spanId": sc.SpanID.String(),
        "sampled": sc.IsSampled(),
    }
}


type logCtx struct {
    time time.Time
    duration time.Duration
    request *http.Request
    response *ResponseLogger
}

func (l logCtx) ElapsedTime() time.Duration {
    return l.duration
}

func (l logCtx) Request() *http.Request {
    return l.request
}

func (l logCtx) RequestTime() time.Time {
    return l.time
}

func (l logCtx) ResponseContentLength() int64 {
    return int64(l.response.size)
}

func (l logCtx) ResponseHeader() http.Header {
    return l.response.ResponseWriter.Header()
}

func (l logCtx) ResponseStatus() int {
    return l.response.status
}

func (l logCtx) ResponseTime() time.Time {
    return l.response.time
}


func sdHook(
    logger *sdlog.Logger,
    logEntry *log.Entry,
    duration time.Duration,
    r *http.Request,
    l *ResponseLogger,
    span *trace.Span,
) {
    if logger == nil {
        return
    }
    payload, _ := json.Marshal(logEntry.Data)
    entry := sdlog.Entry{
        Payload: json.RawMessage(payload),
        HTTPRequest: &sdlog.HTTPRequest{
            Request: r,
            RequestSize: r.ContentLength,
            Status: l.status,
            ResponseSize: int64(l.size),
            Latency: duration,
        },
    }
    if span != nil {
        sc := span.SpanContext()
        entry.Trace = sc.TraceID.String()
        entry.SpanID = sc.SpanID.String()
        entry.TraceSampled = sc.IsSampled()
    }
    logger.Log(entry)
}

func Handler(options Options) func(http.Handler) http.Handler {
    format := propagation.HTTPFormat{ Header: options.TraceHeader }
    client, _ := sdlog.NewClient(context.Background(), "")

    var sdlogger *sdlog.Logger

    if client != nil {
        sdlogger = client.Logger("tracing_log")
    }

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
            var span *trace.Span

            if sc, ok := format.SpanContextFromRequest(r); ok {
                spanData := map[string]interface{}{
                    "parent": toMap(sc),
                }
                if span = trace.FromContext(r.Context()); span != nil {
                    for k, v := range toMap(span.SpanContext()) {
                        spanData[k] = v
                    }
                }
                entry = entry.WithField("span", spanData)
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

            duration := time.Since(start)

            entry = entry.WithFields(log.Fields{
                "responseTime": float64(duration.Nanoseconds()) / 1e6, // ms
                "status": l.status,
                "responseHeaders": filterHeader(l.Header()),
                "responseContentLength": l.size,
            })

            buf := new(bytes.Buffer)
            logCtx := logCtx{
                time: start,
                duration: duration,
                request: r,
                response: l,
            }
            apachelog.CombinedLog.WriteLog(buf, logCtx)

            sdHook(sdlogger, entry, duration, r, l, span)

            entry.Info(buf.String())
        })
    }
}
