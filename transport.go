package tracing

import (
    "context"
    "net/http"
    "github.com/pkg/errors"
    "go.opencensus.io/plugin/ochttp"
    "github.com/openrm/module-tracing-golang/propagation"
    orerrors "github.com/openrm/module-tracing-golang/errors"
)

type statusErrorTransport struct {
    http.RoundTripper
}

func (tp *statusErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    resp, err := tp.RoundTripper.RoundTrip(req)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode >= http.StatusBadRequest {
        return nil, orerrors.WithResponse(
            errors.Errorf("request failed with status %d", resp.StatusCode),
            resp,
        )
    }

    return resp, nil
}

func NewTransport(ctx context.Context) http.RoundTripper {
    return &ochttp.Transport{
        Base: &statusErrorTransport{ http.DefaultTransport },
        Propagation: &propagation.HTTPFormat{
            Header: traceHeader,
        },
    }
}

func NewClient(ctx context.Context) *http.Client {
    return &http.Client{ Transport: NewTransport(ctx) }
}
