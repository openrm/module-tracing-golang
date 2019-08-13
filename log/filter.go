package log

import (
    "net/http"
)

func filterHeader(headers http.Header) map[string]string {
    filtered := make(map[string]string)
    for k, _ := range headers {
        var value string = headers.Get(k)

        for _, re := range filterHeaderPatterns {
            if re.MatchString(k) {
                value = Mask
            }
        }

        filtered[k] = value
    }
    return filtered
}
