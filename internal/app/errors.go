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

package app

import "errors"

var (
	ErrUnknownConfigItemType   = errors.New("unknown item type")
	ErrClusterNotFound         = errors.New("cluster not found")
	ErrAliasAlreadyUsed        = errors.New("alias already in use")
	ErrSourceLocationRequired  = errors.New("source location is required for importing")
	ErrHistoryLocationRequired = errors.New("history location is required")
	ErrHistoryIDRequired       = errors.New("history id is required")
	ErrAliasRequired           = errors.New("alias is required")
	ErrAliasAndIDNotAllowed    = errors.New("alias and id bith specified, only 1 is allowed")
)
