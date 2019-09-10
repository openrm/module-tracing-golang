package tracing

const (
    defaultTraceHeader = "Sentry-Trace"
)

var traceHeader = defaultTraceHeader

func SetTraceHeader(v string) {
    traceHeader = v
}
