package copywriter

import (
	"fmt"
	"os"
	"time"

	"github.com/nullswan/nomi/internal/tools"
)

const exportDirectory = "exports"

// This agent is responsible for exporting the final content
// Capabilities include: Export to file, export to different formats
type exportAgent struct {
	logger  tools.Logger
	project string
}

// TODO(nullswan): add project handling
// TODO(nullswan): add file manager tools
func NewExportAgent(
	logger tools.Logger,
	project string,
) *exportAgent {
	return &exportAgent{
		logger:  logger,
		project: project,
	}
}

func (e *exportAgent) ExportToFile(content string) error {
	fileName := e.project + time.Now().Format("2006-01-02") + ".txt"

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	fmt.Printf("Content exported to %s\n", fileName)
	return nil
}
