package log

import (
    "fmt"
    "mime/multipart"
    "net/http"
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

func parseError(err error) map[string]interface{} {
    errMap := map[string]interface{}{
        "message": err.Error(),
    }

    st := errors.WithStackTrace(err).StackTrace()

    if len(st) > 0 {
        stackTrace := make([]string, len(st))

        for i, frame := range st {
            stackTrace[i] = fmt.Sprintf("%+v", frame)
        }

        errMap["stack"] = stackTrace
    }

    return errMap
}

func parsePanic(err interface{}) map[string]interface{} {
    if err, ok := err.(error); ok {
        return parseError(err)
    }
    if errStr, ok := err.(string); ok {
        return map[string]interface{}{
            "message": errStr,
        }
    }
    return nil
}
