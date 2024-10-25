package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Console interface {
	Exec(context.Context, *Command) (ExecResult, error)
}

type bashConsole struct{}

func NewBashConsole() Console {
	return &bashConsole{}
}

func (c *bashConsole) Exec(
	ctx context.Context,
	cmd *Command,
) (ExecResult, error) {
	command := exec.CommandContext(ctx, cmd.Command, cmd.Args...)
	if cmd.Input != "" {
		command.Stdin = strings.NewReader(cmd.Input)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return ExecResult{}, fmt.Errorf("failed to run command: %w", err)
		}
	}
	return ExecResult{
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
	}, nil
}

type ExecResult struct {
	ExitCode int
	Output   string
	Error    string
}

func (e ExecResult) Success() bool {
	return e.ExitCode == 0
}

type Command struct {
	Command string
	Args    []string
	Input   string
}

func (c *Command) String() string {
	return c.Command + " " + strings.Join(c.Args, " ")
}

func NewCommand(command string, args ...string) *Command {
	return &Command{
		Command: command,
		Args:    args,
	}
}

func (c *Command) WithArgs(args ...string) *Command {
	c.Args = append(c.Args, args...)
	return c
}

func (c *Command) WithInput(input string) *Command {
	c.Input = input
	return c
}
