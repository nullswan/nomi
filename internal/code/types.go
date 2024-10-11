package code

type CodeBlock struct {
	Language string
	Code     string
}

type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Executor interface {
	Execute(code string) ExecutionResult
}
