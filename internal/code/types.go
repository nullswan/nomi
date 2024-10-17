package code

type CodeBlock struct {
	ID          string
	Language    string
	Code        string
	Description string
}

type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Block    CodeBlock
}

type Executor interface {
	Execute(code string) ExecutionResult
}
