package utils

import (
	"regexp"
	"strings"
)

// SurveyFilter a function for passing to AlecAivazis/survey, which will allow wildcards(*) and whitespace to be used for subfilter values
func SurveyFilter(filter string, value string, index int) bool {
	parsedFilter := regexp.MustCompile(`[\s]+`).ReplaceAllString(filter, "*")
	subFilters := strings.Split(parsedFilter, "*")
	for _, s := range subFilters {
		if !strings.Contains(value, s) && s != "" {
			return false
		}
	}
	return true
}
