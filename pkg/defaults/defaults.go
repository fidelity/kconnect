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

package defaults

import (
	"os"
	"path"
)

const (
	RootFolderName  = ".kconnect"
	MaxHistoryItems = 100
	// DefaultUIPageSize specifies the default number of items to display to a user
	DefaultUIPageSize = 10

	UsernameConfigItem = "username"
	PasswordConfigItem = "password"
)

const (
	// PrivateFilePermission is the permission for the files that kconnect writes. The
	// history can contain credentials that were supplied as flags, and the configuration
	// decides where credentials are sent, so only the user should be able to read or
	// change them.
	PrivateFilePermission os.FileMode = 0o600
	// PrivateDirPermission is the permission for the directories that kconnect creates
	PrivateDirPermission os.FileMode = 0o700
)

// WritePrivateFile writes data to a file that only the current user can access. The
// permissions of a file that already exists are tightened as well.
func WritePrivateFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, PrivateFilePermission); err != nil {
		return err
	}

	return os.Chmod(path, PrivateFilePermission)
}

func AppDirectory() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return path.Join(dir, RootFolderName)
}

func HistoryPath() string {
	appDir := AppDirectory()

	return path.Join(appDir, "history.yaml")
}

func ConfigPath() string {
	appDir := AppDirectory()

	return path.Join(appDir, "config.yaml")
}
