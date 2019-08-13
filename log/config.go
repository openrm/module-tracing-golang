package log

import (
    "regexp"
)

const Mask = "[Filtered]"

var (
    excludePatterns = []*regexp.Regexp{}
    filterHeaderPatterns = []*regexp.Regexp{
        regexp.MustCompile(`[Aa]uthorization`),
        regexp.MustCompile(`[Cc]ookie`),
        regexp.MustCompile(`[Ss]et-[Cc]ookie`),
    }
)

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
