package main

import (
	"fmt"
	"strings"

	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Automatically update the application",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		latestVersion, downloadURL, err := getLatestRelease()
		if err != nil {
			fmt.Printf("Failed to fetch latest release: %v\n", err)
			return
		}

		currentVersion := strings.TrimPrefix(
			buildVersion,
			"v",
		) // Remove 'v' prefix
		current, err := semver.Parse(currentVersion)
		if err != nil {
			fmt.Printf("Invalid current version format: %v\n", err)
			return
		}

		latest, err := semver.Parse(latestVersion)
		if err != nil {
			fmt.Printf("Invalid latest version format: %v\n", err)
			return
		}

		if !latest.GT(current) {
			fmt.Println("You are already running the latest version.")
			return
		}

		fmt.Printf("Updating from version %s to %s...\n", current, latest)
		if err := downloadAndReplace(downloadURL); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			return
		}

		fmt.Println("Update successful. Please restart the application.")
	},
}

func getLatestRelease() (string, string, error) {
	url := "https://api.github.com/repos/nullswan/golem/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
			Name               string `json:"name"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	version := strings.TrimPrefix(release.TagName, "v") // Remove 'v' prefix
	assetName := fmt.Sprintf(
		"golem_%s_%s_%s.tar.gz",
		version,
		runtime.GOOS,
		runtime.GOARCH,
	)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("asset %s not found", assetName)
	}

	return version, downloadURL, nil
}

func downloadAndReplace(url string) error {
	// Download the tar.gz asset
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "golem-update-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Extract the tar.gz
	extractDir, err := os.MkdirTemp("", "golem-update")
	if err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}
	defer os.RemoveAll(extractDir)

	if err := extractTarGz(tmpFile.Name(), extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Replace the binary
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	newExePath := filepath.Join(extractDir, binName+"-cli")
	err = os.Rename(newExePath, exePath)
	if err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		default:
			// Skip other file types
		}
	}
	return nil
}
