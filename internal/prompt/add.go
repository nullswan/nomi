package prompts

import (
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"
)

func AddPromptFromURL(url string) (*Prompt, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching the URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"error fetching the URL: received status code %d",
			resp.StatusCode,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var prompt Prompt
	err = yaml.Unmarshal(data, &prompt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	// Validate the prompt
	if err := prompt.Validate(); err != nil {
		return nil, fmt.Errorf("error validating prompt: %v", err)
	}

	// Save the prompt
	if err := prompt.Save(); err != nil {
		return nil, fmt.Errorf("error saving prompt: %v", err)
	}

	return &prompt, nil
}
