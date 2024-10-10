package providers

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/nullswan/golem/internal/config"
	baseprovider "github.com/nullswan/golem/internal/providers/base"
	"github.com/nullswan/golem/internal/providers/ollamaprovider"
	"github.com/nullswan/golem/internal/providers/openaiprovider"
)

func CheckProvider() string {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}

	return "ollama"
}

func LoadTextToTextProvider(
	provider string,
	model string,
) (baseprovider.TextToTextProvider, error) {
	switch provider {
	case "openai":
		oaiConfig := openaiprovider.NewOAIProviderConfig(
			os.Getenv("OPENAI_API_KEY"),
			model,
		)
		return openaiprovider.NewTextToTextProvider(
			oaiConfig,
		)
	case "ollama":
		var cmd *exec.Cmd
		if !ollamaServerIsRunning() {
			var err error
			cmd, err = tryStartOllama()
			if err != nil {
				ollamaOutput := config.GetProgramDirectory() + "/ollama"
				err = backoff.Retry(func() error {
					fmt.Printf(
						"Download ollama to %s\n",
						ollamaOutput,
					)
					return downloadOllama(
						context.TODO(),
						ollamaOutput,
					)
				}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 3))
				if err != nil {
					return nil, fmt.Errorf("error installing ollama: %w", err)
				}
			}
		}
		url := getOllamaURL()

		ollamaConfig := ollamaprovider.NewOlamaProviderConfig(
			url,
			model,
		)
		return ollamaprovider.NewTextToTextProvider(
			ollamaConfig,
			cmd,
		)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func ollamaServerIsRunning() bool {
	defaultURL := "http://localhost:11434"
	req := fmt.Sprintf("%s/health", defaultURL)
	_, err := http.Get(req)
	if err != nil {
		return false
	}
	return true
}

func tryStartOllama() (*exec.Cmd, error) {
	binary := "ollama"
	path, err := exec.LookPath(binary)
	if err == nil {
		cmd := exec.Command(path, "serve")
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("error starting ollama: %w", err)
		}

		fmt.Println("Ollama server started using binary:", path)
		return cmd, nil
	}

	localTarget := config.GetProgramDirectory() + "/ollama"
	if _, err := os.Stat(localTarget); os.IsNotExist(err) {
		fmt.Println("Downloading ollama...")
		err := downloadOllama(context.Background(), localTarget)
		if err != nil {
			return nil, fmt.Errorf("error installing ollama: %w", err)
		}

		if err := os.Chmod(localTarget, 0o755); err != nil {
			return nil, fmt.Errorf(
				"error setting permissions on ollama binary: %w",
				err,
			)
		}

		fmt.Println("Ollama binary downloaded to:", localTarget)

		cmd := exec.Command(localTarget, "serve")
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("error starting ollama: %w", err)
		}

		fmt.Println("Ollama server started using binary:", localTarget)
		return cmd, nil
	}

	return nil, fmt.Errorf("unable to find ollama binary")
}

// this code is part of: https://github.com/redpanda-data/connect/blob/main/internal/impl/ollama/subprocess_unix.go
func downloadOllama(
	ctx context.Context,
	path string,
) error {
	var url string
	const baseURL string = "https://github.com/ollama/ollama/releases/download/v0.3.12/ollama"
	switch runtime.GOOS {
	case "darwin":
		// They ship an universal executable for darwin
		url = baseURL + "-darwin"
	case "linux":
		url = fmt.Sprintf(
			"%s-%s-%s.tgz",
			baseURL,
			runtime.GOOS,
			runtime.GOARCH,
		)
	default:
		return fmt.Errorf(
			"automatic download of ollama is not supported on %s, please download ollama manually",
			runtime.GOOS,
		)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to download ollama binary: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download ollama binary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"failed to download ollama binary: status_code=%d",
			resp.StatusCode,
		)
	}
	var binary io.Reader = resp.Body
	if strings.HasSuffix(url, ".tgz") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf(
				"unable to read tarball for ollama binary download: %w",
				err,
			)
		}
		reader := tar.NewReader(gz)
		for {
			header, err := reader.Next()
			if err == io.EOF {
				return fmt.Errorf(
					"unable to find ollama binary within tarball at %s",
					url,
				)
			} else if err != nil {
				return fmt.Errorf("unable to read tarball at %s: %w", url, err)
			}
			if !header.FileInfo().Mode().IsRegular() ||
				header.Name != "./bin/ollama" {
				continue
			}
			binary = reader
			break
		}
	}

	ollama, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o755)
	if err != nil {
		return fmt.Errorf(
			"unable to create file for ollama binary download: %w",
			err,
		)
	}
	defer ollama.Close()

	_, err = io.Copy(ollama, binary)
	if err != nil {
		return fmt.Errorf(
			"unable to download ollama binary to filesystem: %w",
			err,
		)
	}
	return err
}

func getOllamaURL() string {
	if os.Getenv("OLLAMA_URL") != "" {
		return os.Getenv("OLLAMA_URL")
	}

	return "http://localhost:11434"
}
