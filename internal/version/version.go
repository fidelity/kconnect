package version

import (
	"fmt"
)

var (
	// Version specifies the application version
	Version string

	// BuildDate is the date the CLI was built
	BuildDate string

	// CommitHash is the Git commit hash
	CommitHash string
)

// ToString will convert the version information to a string
func ToString() string {
	return fmt.Sprintf("Version: %s, Build Date: %s, Git Hash: %s", Version, BuildDate, CommitHash)
}
