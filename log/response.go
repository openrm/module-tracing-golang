package log

import (
    "net/http"
)

type responseLogger struct {
    http.ResponseWriter
    status int
    size int
}

func newResponseLogger(w http.ResponseWriter) *responseLogger {
    return &responseLogger{w, http.StatusOK, 0}
}

func (l *responseLogger) WriteHeader(status int) {
    l.status = status
    l.ResponseWriter.WriteHeader(status)
}

func (l *responseLogger) Write(b []byte) (int, error) {
    size, err := l.ResponseWriter.Write(b)
    l.size += size
    return size, err
}
