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
)

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
