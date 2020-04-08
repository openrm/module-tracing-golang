module github.com/openrm/module-tracing-golang

go 1.12

require (
	cloud.google.com/go v0.43.0
	cloud.google.com/go/logging v1.0.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.7
	github.com/getsentry/sentry-go v0.2.1
	github.com/google/uuid v1.1.1
	github.com/lestrrat-go/apache-logformat v2.0.4+incompatible
	github.com/lestrrat-go/strftime v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	go.opencensus.io v0.22.1
)
