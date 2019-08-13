package tracing

import (
    "context"
    "net/http"
)

type tracingTransport struct {
    *http.Transport
    span *Span
}

func (tp *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    if tp.span != nil {
        req.Header.Set(traceHeader, tp.span.Serialize())
    }
    return tp.Transport.RoundTrip(req)
}

func NewTransport(ctx context.Context) http.RoundTripper {
    defaultTransport := http.DefaultTransport.(*http.Transport)
    tp := tracingTransport{ Transport: defaultTransport }

    if sp, ok := ctx.Value(SpanContextKey).(*Span); ok {
        tp.span = sp
    }

    return &tp
}

func NewClient(ctx context.Context) *http.Client {
    return &http.Client{ Transport: NewTransport(ctx) }
}
