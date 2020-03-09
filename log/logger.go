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

type contextLogger struct {
    *log.Entry
    hub *sentry.Hub
}

func NewLogger(entry *log.Entry) Logger {
    return &contextLogger{ Entry: entry }
}

func (l *contextLogger) withError(err error) contextLogger {
    return contextLogger{ l.Entry.WithField(errorKey, parseError(err)), l.hub }
}

func (l *contextLogger) WithError(err error) *log.Entry {
    return l.withError(err).Entry
}

func (l *contextLogger) Report(err error) {
    if l.hub != nil {
        id := l.hub.CaptureException(err)
        if id != nil {
            l.WithError(err).Errorf("tracing: error reported: %s", *id)
        }
    }
}

func MustGetLogger(ctx context.Context) Logger {
    logger, ok := ctx.Value(LoggerContextKey).(*contextLogger)

    if !ok {
        panic("tracing: could not get logger from context")
    }

    return &contextLogger{
        Entry: logger.Entry,
        hub: sentry.GetHubFromContext(ctx),
    }
}

func GetLogger(ctx context.Context) Logger {
    logger, ok := ctx.Value(LoggerContextKey).(*contextLogger)

    if ok {
        return &contextLogger{
            Entry: logger.Entry,
            hub: sentry.GetHubFromContext(ctx),
        }
    }

    return nil
}
