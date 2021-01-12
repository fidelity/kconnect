/*
Copyright 2020 The kconnect Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
