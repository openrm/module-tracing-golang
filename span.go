package tracing

import (
    "fmt"
    "strings"
    "github.com/google/uuid"
)

type Span struct {
    TraceId string
    SpanId string
    Parent *Span
}

func (s *Span) Serialize() string {
    return fmt.Sprintf("%s-%s", s.TraceId, s.SpanId)
}

func genSpanId() string {
    id := uuid.New()
    hex := strings.Join(strings.Split(id.String(), "-"), "")
    return hex[:16]
}

func newSpanFromParent(p Span) Span {
    return Span{ Parent: &p, TraceId: p.TraceId, SpanId: genSpanId() }
}
