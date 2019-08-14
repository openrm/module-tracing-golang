package errors

import (
    "runtime"
    "github.com/pkg/errors"
)

type nestedStack interface {
    error
    stackTracer
    causer
}

type stackTracer interface {
    StackTrace() errors.StackTrace
}

type causer interface {
    Cause() error
}

func newStack() []uintptr {
    pc := make([]uintptr, 128)
    n := runtime.Callers(1, pc)
    return pc[:n]
}

// recursively recover and track the stack frames attached in the error
func extractStackTrace(err error, seen ...uintptr) []uintptr {
    if err, ok := err.(stackTracer); ok {
        st := err.StackTrace()

        var stackTrace []uintptr
        unique := len(seen) == 0

        for i := 0; i < len(st); i++ {
            frame := uintptr(st[len(st) - 1 - i])
            if unique || len(seen) > i && seen[len(seen) - 1 - i] != frame {
                unique = true
                stackTrace = append([]uintptr{frame}, stackTrace...)
            }
        }

        if err, ok := err.(nestedStack); ok {
            return append(extractStackTrace(err.Cause(), append(stackTrace, seen...)...), stackTrace...)
        }

        return stackTrace
    }

    if len(seen) == 0 {
        return newStack()
    }

    return []uintptr{}
}

func fromPC(pcs []uintptr) []errors.Frame {
    fs := make([]errors.Frame, len(pcs))
    for i, pc := range pcs {
        fs[i] = errors.Frame(pc)
    }
    return fs
}

type flattened struct {
    error
    stack []uintptr
}

func (f flattened) StackTrace() errors.StackTrace {
    return fromPC(f.stack)
}

func WithStackTrace(err error) flattened {
    return flattened{ err, extractStackTrace(err) }
}

func NewStackTrace() errors.StackTrace {
    stack := newStack()
    return fromPC(stack)
}
