package tracing

import (
    "log"
    "contrib.go.opencensus.io/exporter/stackdriver"
    "contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
    "go.opencensus.io/trace"
)

func InitTracer() {
    exporter, err := stackdriver.NewExporter(stackdriver.Options{
        MonitoredResource: monitoredresource.Autodetect(),
    })
    if err != nil {
        log.Fatal(err)
    }
    trace.RegisterExporter(exporter)
    trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}
