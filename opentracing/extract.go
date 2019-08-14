package opentracing

import (
    "regexp"
)

var traceParentPattern = regexp.MustCompile(`^[ \t]*([0-9a-f]{32})?-?([0-9a-f]{16})?-?([01])?[ \t]*$`)

func extractSpan(v string) *Span {
    if matches := traceParentPattern.FindStringSubmatch(v); len(matches) > 2 {
        return &Span{ TraceId: matches[1], SpanId: matches[2] }
    }
    return nil
}

func NewFromTraceParent(v string) *Span {
    if parent := extractSpan(v); parent != nil {
        span := NewFromParent(*parent)
        return &span
    }
    return nil
}
