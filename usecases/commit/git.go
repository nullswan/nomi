package commit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nullswan/nomi/internal/tools"
)

func checkGitRepository(ctx context.Context, console tools.Console) error {
	cmd := tools.NewCommand("git", "rev-parse", "--is-inside-work-tree")
	result, err := console.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to check git repository: %w", err)
	}
	if !result.Success() {
		return errors.New("not a git repository")
	}
	return nil
}

func stashChanges(ctx context.Context, console tools.Console) error {
	timestamp := time.Now().Format("20060102T150405")
	stashName := "nomi-stash-" + timestamp
	cmd := tools.NewCommand(
		"git",
		"stash",
		"push",
		"--include-untracked",
		"--message",
		stashName,
	)
	result, err := console.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to stash changes: %w", err)
	}
	if !result.Success() {
		if result.Error != "" {
			return fmt.Errorf("failed to stash changes: %s", result.Error)
		}
		if result.Output != "" {
			return fmt.Errorf("failed to stash changes: %s", result.Output)
		}
		return errors.New("failed to stash changes  and received no output")
	}

	// Extract stash reference from the output
	stashRef := ""
	lines := strings.Split(result.Output, "\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], ":")
		if len(parts) > 0 {
			stashRef = strings.TrimSpace(parts[0])
		}
	}
	if stashRef == "" {
		return errors.New("unable to retrieve stash reference")
	}

	return nil
}

func getStashDiff(
	ctx context.Context,
	console tools.Console,
) (string, error) {
	cmd := tools.NewCommand(
		"git",
		"stash",
		"show",
		"--include-untracked",
		"--patch",
		"stash@{0}",
	)
	result, err := console.Exec(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to show stash diff: %w", err)
	}
	if !result.Success() {
		return "", errors.New("failed to show stash diff")
	}
	return result.Output, nil
}

func unstashChanges(
	ctx context.Context,
	console tools.Console,
) error {
	cmd := tools.NewCommand("git", "stash", "apply", "stash@{0}")
	result, err := console.Exec(ctx, cmd)
	if err != nil || !result.Success() {
		return errors.New("failed to unstash changes")
	}

	cmd = tools.NewCommand("git", "reset")
	result, err = console.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to reset changes: %w", err)
	}

	if !result.Success() {
		return errors.New("failed to reset changes")
	}

	return nil
}

func deleteStash(
	ctx context.Context,
	console tools.Console,
) error {
	cmd := tools.NewCommand("git", "stash", "drop", "stash@{0}")
	result, err := console.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete stash: %w", err)
	}
	if !result.Success() {
		return errors.New("failed to delete stash")
	}
	return nil
}
