package tracing

import (
    "io/ioutil"
    "context"
    "net/http"
    "github.com/pkg/errors"
    "github.com/openrm/module-tracing-golang/opentracing"
)

type tracingTransport struct {
    *http.Transport
    span *opentracing.Span
}

func (tp *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    if tp.span != nil {
        req.Header.Set(traceHeader, tp.span.Serialize())
    }

    resp, err := tp.Transport.RoundTrip(req)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode >= http.StatusBadRequest {
        if body, err := ioutil.ReadAll(resp.Body); err == nil {
            return nil, errors.Errorf("transport: %d - \"%s\"", resp.StatusCode, body)
        } else {
            return nil, errors.New("transport: error response")
        }
    }

    return resp, err
}

func NewTransport(ctx context.Context) http.RoundTripper {
    defaultTransport := http.DefaultTransport.(*http.Transport)
    tp := tracingTransport{ Transport: defaultTransport }

    if sp, ok := ctx.Value(SpanContextKey).(*opentracing.Span); ok {
        tp.span = sp
    }

    return &tp
}

func NewClient(ctx context.Context) *http.Client {
    return &http.Client{ Transport: NewTransport(ctx) }
}
