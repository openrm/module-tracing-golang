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
        pc := make([]uintptr, 128)
        n := runtime.Callers(1, pc)
        return pc[:n]
    }

    return []uintptr{}
}

type flattened struct {
    error
    stack []uintptr
}

func (f flattened) StackTrace() errors.StackTrace {
	fs := make([]errors.Frame, len(f.stack))
	for i := 0; i < len(fs); i++ {
		fs[i] = errors.Frame(f.stack[i])
	}
	return fs
}

func WithStackTrace(err error) flattened {
    return flattened{ err, extractStackTrace(err) }
}
