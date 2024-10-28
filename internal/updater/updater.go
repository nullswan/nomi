package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/mholt/archiver/v3"
)

type Config struct {
	Repository     string
	CurrentVersion string
	BinaryName     string
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
}

type Updater struct {
	config Config
}

func New(config Config) *Updater {
	return &Updater{config: config}
}

func (u *Updater) Update() error {
	fmt.Println("Checking for updates...")

	latestVersion, downloadURL, archiveType, err := u.getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	currentVersion := strings.TrimPrefix(u.config.CurrentVersion, "v")
	current, err := semver.Parse(currentVersion)
	if err != nil {
		return fmt.Errorf("invalid current version format: %w", err)
	}

	latest, err := semver.Parse(latestVersion)
	if err != nil {
		return fmt.Errorf("invalid latest version format: %w", err)
	}

	if !latest.GT(current) {
		fmt.Println("You are already running the latest version.")
		return nil
	}

	fmt.Printf("Updating from version %s to %s...\n", current, latest)
	downloadPath, err := u.downloadAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(downloadPath)

	if err := u.installUpdate(downloadPath, archiveType); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("Update successful. Please restart the application.")
	return nil
}

func (u *Updater) getLatestRelease() (string, string, string, error) {
	release, err := fetchLatestRelease(u.config.Repository)
	if err != nil {
		return "", "", "", err
	}

	version := strings.TrimPrefix(release.TagName, "v")
	downloadURL, archiveType, err := getDownloadDetails(
		release.Assets,
		u.config.BinaryName,
	)
	if err != nil {
		return "", "", "", err
	}

	return version, downloadURL, archiveType, nil
}

func fetchLatestRelease(repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

func getDownloadDetails(
	assets []Asset,
	binaryName string,
) (string, string, error) {
	var downloadURL, archiveType string

	var expectedName string
	if runtime.GOOS == "windows" {
		expectedName = fmt.Sprintf(
			"%s-cli-windows-%s.zip",
			binaryName,
			mapArch(runtime.GOARCH),
		)
		archiveType = "zip"
	} else {
		expectedName = fmt.Sprintf("%s-cli-%s-%s", binaryName, runtime.GOOS, mapArch(runtime.GOARCH))
		archiveType = "binary"
	}

	for _, asset := range assets {
		if asset.Name == expectedName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf(
			"no suitable asset found for OS: %s and ARCH: %s",
			runtime.GOOS,
			runtime.GOARCH,
		)
	}

	return downloadURL, archiveType, nil
}

func (u *Updater) downloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to download asset, status: %s",
			resp.Status,
		)
	}

	ext := getFileExtension(url)
	tmpFile, err := os.CreateTemp("", "nomi-update-*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save asset: %w", err)
	}

	return tmpFile.Name(), nil
}

func getFileExtension(url string) string {
	return filepath.Ext(url)
}

func (u *Updater) installUpdate(downloadPath, archiveType string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	var newExePath string
	switch archiveType {
	case "zip":
		extractDir := filepath.Dir(downloadPath)
		if err := extractAsset(downloadPath, extractDir); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
		newExePath = filepath.Join(extractDir, u.config.BinaryName+".exe")
	case "binary":
		newExePath = downloadPath
		if runtime.GOOS != "windows" {
			if err := os.Chmod(newExePath, 0o755); err != nil {
				return fmt.Errorf(
					"failed to set executable permissions: %w",
					err,
				)
			}
		}
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}

	return replaceBinary(newExePath, exePath)
}

func extractAsset(downloadPath, extractDir string) error {
	return archiver.Unarchive(downloadPath, extractDir)
}

func mapArch(goArch string) string {
	switch goArch {
	case "amd64":
		return "amd64"
	case "386":
		return "386"
	case "arm64":
		return "arm64"
	default:
		return goArch
	}
}

func replaceBinary(newPath, oldPath string) error {
	backupPath := oldPath + ".backup"
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old binary: %w", err)
	}
	defer os.Remove(backupPath)

	if runtime.GOOS == "windows" {
		return os.Rename(newPath, oldPath)
	}

	return os.Rename(newPath, oldPath)
}
