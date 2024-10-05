package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetHomeDir returns the user's home directory.
func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

// GetDataDir returns the application's data directory, e.g., ~/.ai
func GetDataDir() string {
	homeDir := GetHomeDir()
	dataDir := filepath.Join(homeDir, configDir)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.MkdirAll(dataDir, 0o755)
	}
	return dataDir
}

// GetDataSubdir returns the specified subdirectory under the data directory.
func GetDataSubdir(subdir string) string {
	dataDir := GetDataDir()
	dir := filepath.Join(dataDir, subdir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0o755)
	}
	return dir
}
