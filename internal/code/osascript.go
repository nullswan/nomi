package code

import (
	"os/exec"
	"strings"
)

type OsascriptExecutor struct{}

func (oe *OsascriptExecutor) Execute(code string) ExecutionResult {
	cmd := exec.Command("osascript", "-e", code)
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

func initOsascriptExecutor() {
	registerExecutor("osascript", &OsascriptExecutor{})
}
