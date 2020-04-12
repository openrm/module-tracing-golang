package tracing

import (
    "log"
    "contrib.go.opencensus.io/exporter/stackdriver"
    "contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
    "go.opencensus.io/trace"
)

func InitTracer() func() {
    exporter, err := stackdriver.NewExporter(stackdriver.Options{
        MonitoredResource: monitoredresource.Autodetect(),
    })
    if err != nil {
        log.Fatal(err)
    }
    exporter.StartMetricsExporter()
    trace.RegisterExporter(exporter)
    trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
    return func() {
        exporter.Flush()
        exporter.StopMetricsExporter()
        trace.UnregisterExporter(exporter)
    }
}
