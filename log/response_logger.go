package log

import (
    "net/http"
)

type ResponseLogger struct {
    http.ResponseWriter
    Status int
    Size int
    Error error
}

func NewResponseLogger(w http.ResponseWriter) *ResponseLogger {
    return &ResponseLogger{w, http.StatusOK, 0, nil}
}

func (l *ResponseLogger) WriteHeader(status int) {
    l.Status = status
    l.ResponseWriter.WriteHeader(status)
}

func (l *ResponseLogger) Write(b []byte) (int, error) {
    size, err := l.ResponseWriter.Write(b)
    l.Size += size
    return size, err
}

func (l *ResponseLogger) WriteError(err error) {
    l.Error = err
}
