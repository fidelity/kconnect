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

package errors

import (
	"fmt"
	"strings"
)

type ValidationFailed struct {
	validationErrors []string
}

func (e *ValidationFailed) Error() string {
	errorList := strings.Join(e.validationErrors[:], "\n")
	return fmt.Sprintf("Validation failed with the following:\n%s", errorList)
}

func (e *ValidationFailed) Failures() []string {
	return e.validationErrors
}

func (e *ValidationFailed) AddFailure(failureText string) {
	e.validationErrors = append(e.validationErrors, failureText)
}

func (e *ValidationFailed) setup() {
	if e.validationErrors != nil {
		return
	}
	e.validationErrors = []string{}
}

func IsValidationFailed(err error) bool {
	if _, ok := err.(*ValidationFailed); ok {
		return true
	}

	return false
}
