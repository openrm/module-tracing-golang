package errors

import (
    "strings"
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

func newStack() errors.StackTrace {
    pc := make([]uintptr, 128)
    n := runtime.Callers(1, pc)
    return fromPC(pc[:n])
}

func pc(f errors.Frame) uintptr {
    return uintptr(f) - 1
}

func fromPC(pcs []uintptr) errors.StackTrace {
    fs := make([]errors.Frame, len(pcs))
    for i, pc := range pcs {
        fs[i] = errors.Frame(pc)
    }
    return fs
}

// recursively recover and track the stack frames attached in the error
func extractStackTrace(err error) errors.StackTrace {
    if err, ok := err.(stackTracer); ok {
        stack := err.StackTrace()

        if err, ok := err.(causer); ok {
            return append(extractStackTrace(err.Cause()), stack...)
        }

        return stack
    }

    if err, ok := err.(causer); ok {
        return extractStackTrace(err.Cause())
    }

    return errors.StackTrace{}
}

func ExtractStackTrace(err error) errors.StackTrace {
    return extractStackTrace(err)
}

type flattened struct {
    error
    stack errors.StackTrace
}

func (f flattened) StackTrace() errors.StackTrace {
    return filterFrame(f.stack)
}

func module(name string) string {
    if i := strings.LastIndex(name, "."); i > -1 {
        return name[:i]
    }
    return ""
}

const (
    identity = "github.com/openrm/module-tracing-golang/errors"
    tracingLogPath = "github.com/openrm/module-tracing-golang/log"
    sentrySdkPath = "github.com/getsentry/sentry-go"
)

func filterFrame(st errors.StackTrace) errors.StackTrace {
    stack := make(errors.StackTrace, 0, len(st))
    for _, frame := range st {
        fn := runtime.FuncForPC(pc(frame))

        if fn == nil {
            continue
        }

        mod := module(fn.Name())

        if mod == "runtime" {
            continue
        }

        if strings.HasPrefix(mod, identity) || strings.HasPrefix(mod, tracingLogPath) {
            continue
        }

        if strings.HasPrefix(mod, sentrySdkPath) {
            continue
        }

        stack = append(stack, frame)
    }
    return stack
}

func WithStackTrace(err error) error {
    if _, ok := err.(stackTracer); ok {
        return flattened{ err, extractStackTrace(err) }
    }
    return flattened{ err, newStack() }
}

func NewStackTrace() errors.StackTrace {
    return filterFrame(newStack())
}
