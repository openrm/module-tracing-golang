package tracing

import (
    "strings"
    "github.com/getsentry/sentry-go"
    "github.com/openrm/module-tracing-golang/opentracing"
)

const (
    integrationName = "Tracing"
    integrationContextNamespace = "trace"
)

type tracingIntegration struct {}

var Integration sentry.Integration = new(tracingIntegration)

func (ti *tracingIntegration) Name() string {
    return integrationName
}

func (ti *tracingIntegration) SetupOnce(client *sentry.Client) {
    client.AddEventProcessor(ti.processor)
}

func (ti *tracingIntegration) processor(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    if event.Contexts == nil {
        event.Contexts = make(map[string]interface{})
    }

    if sp := ti.extractSpan(event.Request); sp != nil {
        span := *sp
        event.Contexts[integrationContextNamespace] = span.JSON(false)
    }

    return event
}

func (ti *tracingIntegration) extractSpan(r sentry.Request) *opentracing.Span {
    var traceParent string

    for k, v := range r.Headers {
        if strings.ToLower(k) == strings.ToLower(traceHeader) {
            traceParent = v
        }
    }

    if traceParent != "" {
        if sp := opentracing.NewFromTraceParent(traceParent); sp != nil {
            return sp
        }
    }

    return nil
}
