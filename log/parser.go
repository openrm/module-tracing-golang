package log

import (
    "mime/multipart"
    "net/http"
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
