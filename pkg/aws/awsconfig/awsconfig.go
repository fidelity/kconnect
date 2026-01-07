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

package awsconfig

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// LocateConfigFile - Taken from saml2aws
func LocateConfigFile() (string, error) {
	filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")

	if filename != "" {
		return filename, nil
	}

	var name string

	var err error

	if runtime.GOOS == "windows" {
		name = path.Join(os.Getenv("USERPROFILE"), ".aws", "credentials")
	} else {
		name, err = homedir.Expand("~/.aws/credentials")
		if err != nil {
			return "", err
		}
	}

	// is the filename a symlink?
	name, err = resolveSymlink(name)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve symlink")
	}

	return name, nil
}

func resolveSymlink(filename string) (string, error) {
	sympath, err := filepath.EvalSymlinks(filename)

	// return the un modified filename
	if os.IsNotExist(err) {
		return filename, nil
	}

	if err != nil {
		return "", err
	}

	return sympath, nil
}
