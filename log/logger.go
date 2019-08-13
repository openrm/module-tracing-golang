package log

import (
    "context"
    "github.com/getsentry/sentry-go"
    log "github.com/sirupsen/logrus"
)

type Logger interface {
    log.FieldLogger
    reporter
}

type reporter interface {
    Report(err error)
}

type logger struct {
    *log.Entry
    hub *sentry.Hub
}

func NewLogger(entry *log.Entry) Logger {
    return &logger{ Entry: entry }
}

func (l *logger) withError(err error) logger {
    return logger{ l.Entry.WithField(errorKey, parseError(err)), l.hub }
}

func (l *logger) WithError(err error) *log.Entry {
    return l.withError(err).Entry
}

func (l *logger) Report(err error) {
    id := l.hub.CaptureException(err)
    if id != nil {
        l.WithError(err).Errorf("tracing: error reported: %s", *id)
    }
}

func MustGetLogger(ctx context.Context) Logger {
    logger, ok := ctx.Value(LoggerContextKey).(*logger)

    if !ok {
        panic("tracing: could not get logger from context")
    }

    logger.hub = sentry.GetHubFromContext(ctx)

    return logger
}

func GetLogger(ctx context.Context) Logger {
    logger, ok := ctx.Value(LoggerContextKey).(*logger)

    if ok {
        logger.hub = sentry.GetHubFromContext(ctx)
        return logger
    }

    return nil
}
