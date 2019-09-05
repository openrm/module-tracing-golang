package log

import (
    "fmt"
    "mime/multipart"
    "net/http"
    pkgerrors "github.com/pkg/errors"
    "github.com/openrm/module-tracing-golang/errors"
)

func parseMultipartForm(form *multipart.Form) map[string]interface{} {
    parsedForm := make(map[string]interface{}, len(form.File))

    // the content of form.Value is already set in http.Request.Form
    for k, vs := range form.File {
        files := make([]map[string]interface{}, len(vs))
        for i, f := range vs {
            files[i] = map[string]interface{}{
                "filename": f.Filename,
                "headers": f.Header,
                "size": f.Size,
            }
        }
        parsedForm[k] = files
    }

    return parsedForm
}

func parseCookies(cookies []*http.Cookie) []string {
    serialized := make([]string, len(cookies))
    for i, cookie := range cookies {
        serialized[i] = cookie.String()
    }
    return serialized
}

func sprintStack(st pkgerrors.StackTrace) []string {
    stackTrace := make([]string, len(st))

    for i, frame := range st {
        stackTrace[i] = fmt.Sprintf("%+v", frame)
    }

    return stackTrace
}

func parseError(err error) map[string]interface{} {
    errMap := map[string]interface{}{
        "message": err.Error(),
    }

    st := errors.ExtractStackTrace(err)

    if len(st) == 0 {
        st = errors.NewStackTrace()
    }

    errMap["stack"] = sprintStack(st)

    return errMap
}

func parsePanic(err interface{}) map[string]interface{} {
    if err, ok := err.(error); ok {
        return parseError(err)
    }
    return map[string]interface{}{
        "message": err,
        "stack": sprintStack(errors.NewStackTrace()),
    }
}
