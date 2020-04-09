package log

import (
    "regexp"
    log "github.com/sirupsen/logrus"
)

var globalLogger log.FieldLogger = log.New()

func SetLogger(logger log.FieldLogger) {
    globalLogger = logger
}

const Mask = "[Filtered]"

var (
    excludePatterns = []*regexp.Regexp{}
    filterHeaderPatterns = []*regexp.Regexp{
        regexp.MustCompile(`[Aa]uthorization`),
        regexp.MustCompile(`[Cc]ookie`),
        regexp.MustCompile(`[Ss]et-[Cc]ookie`),
    }
)

func GetExcludePatterns() []*regexp.Regexp {
    return excludePatterns
}

func SetExcludePatterns(exprs ...string) {
    for _, expr := range exprs {
        re, err := regexp.Compile(expr)
        if err == nil {
            excludePatterns = append(excludePatterns, re)
        }
    }
}

func SetFilterHeaderPatterns(exprs ...string) {
    for _, expr := range exprs {
        re, err := regexp.Compile(expr)
        if err == nil {
            filterHeaderPatterns = append(excludePatterns, re)
        }
    }
}

var (
    serviceName string
    serviceVersion string
)

func SetServiceContext(name, version string) {
    serviceName = name
    serviceVersion = version
}
