package code

import (
	"os/exec"
	"strings"
)

type BashExecutor struct{}

func (be *BashExecutor) Execute(code string) ExecutionResult {
	cmd := exec.Command("bash", "-c", code)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecutionResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

func initBashExecutor() {
	registerExecutor("bash", &BashExecutor{})
}
