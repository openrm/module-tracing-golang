package errors

import (
    "io/ioutil"
    "net/http"
    "net/url"
    "encoding/json"
)

type ResponseError interface {
    Error() string
    StatusCode() int
    Body() interface{}
}

func WithResponse(err error, resp *http.Response) error {
    if err == nil {
        return nil
    }
    var body interface{} = "<body unavailable>"
    if data, rerr := ioutil.ReadAll(resp.Body); rerr == nil {
        if err := json.Unmarshal(data, &body); err != nil {
            body = string(data)
        }
    }
    return &withResponse{ err, resp.StatusCode, body }
}

func ExtractResponseError(err error) ResponseError {
    return extractResponseError(err)
}

func extractResponseError(err error) ResponseError {
    if err, ok := err.(*withResponse); ok {
        return err
    }
    if err, ok := err.(*url.Error); ok {
        // http.(*Client).Do() usually wraps an error in this form
        return extractResponseError(err.Err)
    }
    if err, ok := err.(causer); ok {
        return extractResponseError(err.Cause())
    }
    return nil
}

type withResponse struct {
    error
    status int
    body interface{}
}

func (w *withResponse) Cause() error {
    return w.error
}

func (w *withResponse) StatusCode() int {
    return w.status
}

func (w *withResponse) Body() interface{} {
    return w.body
}
