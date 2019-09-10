package tracing

import (
    "github.com/getsentry/sentry-go"
    "go.opencensus.io/trace"
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

    if sp, ok := event.Extra[sentrySpanKey]; ok {
        if span, ok := sp.(*trace.Span); ok {
            var parent map[string]interface{}

            if psc, ok := event.Extra[sentryParentSpanContextKey]; ok {
                if psc, ok := psc.(trace.SpanContext); ok {
                    parent = map[string]interface{}{
                        "trace_id": psc.TraceID.String(),
                        "span_id": psc.SpanID.String(),
                    }
                }
            }

            sc := span.SpanContext()
            event.Contexts[integrationContextNamespace] = map[string]interface{}{
                "parent": parent,
                "trace_id": sc.TraceID.String(),
                "span_id": sc.SpanID.String(),
            }
        }
    }

    return event
}
