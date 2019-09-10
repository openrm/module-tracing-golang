package propagation

import (
    "fmt"
    "regexp"
    "strings"
    "net/http"
    "encoding/binary"
    "encoding/hex"
    "go.opencensus.io/trace"
)

var tracePattern = regexp.MustCompile(`^[ \t]*([0-9a-f]{32})?-?([0-9a-f]{16})?-?([01])?[ \t]*$`)

type HTTPFormat struct {
    Header string
}

func parseHeader(v string) (tid trace.TraceID, sid trace.SpanID, opts trace.TraceOptions, ok bool) {
    if matches := tracePattern.FindStringSubmatch(v); len(matches) > 2 {
        var buf []byte
        var err error

        if buf, err = hex.DecodeString(matches[1]); err != nil {
            return
        }
        copy(tid[:], buf)

        if buf, err = hex.DecodeString(matches[2]); err != nil {
            return
        }
        copy(sid[:], buf)

        if matches[3] == "1" {
            opts = trace.TraceOptions(1)
        } else {
            opts = trace.TraceOptions(0)
        }

        ok = true
        return
    }
    return
}

func (f *HTTPFormat) spanContextFromString(v string) (trace.SpanContext, bool) {
    if tid, sid, opts, ok := parseHeader(v); ok {
        return trace.SpanContext{
            TraceID: tid,
            SpanID: sid,
            TraceOptions: opts,
        }, true
    }
    return trace.SpanContext{}, false
}

func (f *HTTPFormat) SpanContextFromHeaders(headers map[string]string) (trace.SpanContext, bool) {
    for k, v := range headers {
        if strings.ToLower(k) == strings.ToLower(f.Header) {
            return f.spanContextFromString(v)
        }
    }
    return trace.SpanContext{}, false
}

func (f *HTTPFormat) SpanContextFromRequest(r *http.Request) (trace.SpanContext, bool) {
    if str := r.Header.Get(f.Header); str != "" {
        return f.spanContextFromString(str)
    }
    return trace.SpanContext{}, false
}

func serialize(sc trace.SpanContext) string {
    tid := hex.EncodeToString(sc.TraceID[:])
    sid := binary.BigEndian.Uint64(sc.SpanID[:])
    return fmt.Sprintf("%s-%s-%d", tid, sid, int64(sc.TraceOptions))
}

func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, r *http.Request) {
    r.Header.Set(f.Header, serialize(sc))
}
