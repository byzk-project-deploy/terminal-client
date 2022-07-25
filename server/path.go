package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/byzk-project-deploy/terminal-client/user"
)

var (
	currentPath string
)

func init() {
	wdDir, err := os.Getwd()
	if err == nil {
		currentPath = wdDir
		return
	}

	currentPath = user.HomeDir()
}

func CurrentPath() string {
	return currentPath
}

func CurrentPathChange(p string) {
	currentPath = p
}

func JoinPath(p string) string {
	p = strings.TrimSpace(p)
	if p[0] == '/' {
		return p
	}

	return filepath.Join(currentPath, p)
}
