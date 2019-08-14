package log

import (
    "net/http"
)

type ResponseLogger struct {
    http.ResponseWriter
    status int
    size int
    err error
    extra map[interface{}]interface{}
}

func NewResponseLogger(w http.ResponseWriter) *ResponseLogger {
    return &ResponseLogger{w, http.StatusOK, 0, nil, make(map[interface{}]interface{})}
}

func (l *ResponseLogger) WriteHeader(status int) {
    l.status = status
    l.ResponseWriter.WriteHeader(status)
}

func (l *ResponseLogger) Write(b []byte) (int, error) {
    size, err := l.ResponseWriter.Write(b)
    l.size += size
    return size, err
}

func (l *ResponseLogger) WriteError(err error) {
    l.err = err
}

func (l *ResponseLogger) WriteExtra(k, v interface{}) {
    l.extra[k] = v
}

func (l *ResponseLogger) getExtra(k interface{}) interface{} {
    return l.extra[k]
}
