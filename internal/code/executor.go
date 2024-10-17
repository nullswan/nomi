package code

import (
	"fmt"
	"runtime"
	"sync"
)

var executors = make(map[string]Executor)
var onceExecutorRegistration sync.Once

func registerExecutor(language string, executor Executor) {
	executors[language] = executor
}

func ExecuteCodeBlock(block CodeBlock) ExecutionResult {
	onceExecutorRegistration.Do(
		func() {
			initBashExecutor()
			initPythonExecutor()
			initOsascriptExecutor()
		},
	)

	executor, ok := executors[block.Language]
	if !ok {
		return ExecutionResult{
			Stderr:   fmt.Sprintf("Unsupported language: %s", block.Language),
			ExitCode: 1,
		}
	}

	if block.Language == "osascript" && runtime.GOOS != "darwin" {
		return ExecutionResult{
			Stderr:   "Osascript is only supported on macOS",
			ExitCode: 1,
		}
	}

	return executor.Execute(block.Code)
}
