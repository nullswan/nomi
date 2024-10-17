package code

import (
	"fmt"
	"strings"
)

func FormatExecutionResultForLLM(results []ExecutionResult) string {
	var sections []string

	for i, r := range results {
		var sectionParts []string

		sectionParts = append(
			sectionParts,
			fmt.Sprintf("--- Execution Result %d ---", i+1),
		)

		if r.Stderr != "" {
			sectionParts = append(
				sectionParts,
				"Error:\n"+r.Stderr,
			)
		}

		if r.Stdout != "" {
			sectionParts = append(
				sectionParts,
				"Output:\n"+r.Stdout,
			)
		}

		if r.ExitCode != 0 {
			sectionParts = append(
				sectionParts,
				fmt.Sprintf("Exit Code: %d", r.ExitCode),
			)
		}

		if len(sectionParts) > 1 { // If there's any content besides the header
			sections = append(sections, strings.Join(sectionParts, "\n\n"))
		}
	}

	return strings.Join(sections, "\n\n"+strings.Repeat("-", 40)+"\n\n")
}
