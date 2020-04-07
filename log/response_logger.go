package log

import (
    "time"
    "net/http"
)

type ResponseLogger struct {
    http.ResponseWriter
    status int
    size int
    err error
    time time.Time
    extra map[interface{}]interface{}
}

func NewResponseLogger(w http.ResponseWriter) *ResponseLogger {
    return &ResponseLogger{w, http.StatusOK, 0, nil, time.Time{}, make(map[interface{}]interface{})}
}

func (l *ResponseLogger) WriteHeader(status int) {
    l.time = time.Now()
    l.status = status
    l.ResponseWriter.WriteHeader(status)
}

func (l *ResponseLogger) Write(b []byte) (int, error) {
    l.time = time.Now()
    size, err := l.ResponseWriter.Write(b)
    l.size += size
    return size, err
}

func (l *ResponseLogger) WriteError(err error) {
    l.time = time.Now()
    l.err = err
}

func (l *ResponseLogger) WriteExtra(k, v interface{}) {
    l.extra[k] = v
}

func (l *ResponseLogger) getExtra(k interface{}) interface{} {
    return l.extra[k]
}
