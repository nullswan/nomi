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

// GetProgramDirectory returns the application's data directory
// e.g., ~/.golem on Unix systems.
func GetProgramDirectory() string {
	homeDir := GetHomeDir()
	dataDir := filepath.Join(homeDir, configDir)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.MkdirAll(dataDir, 0o755)
	}
	return dataDir
}

// GetModuleDirectory returns the specified subdirectory under the data directory.
func GetModuleDirectory(subdir string) string {
	dataDir := GetProgramDirectory()
	dir := filepath.Join(dataDir, subdir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0o755)
	}
	return dir
}
