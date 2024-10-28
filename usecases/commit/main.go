package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

// TODO(nullswan): Handle stash reference correctly to avoid any TOCTOU issues.
// TODO(nullswan): Add memory on the commit plan, preference, commonly used prefix, scopes, modules, and components.

func OnStart(
	ctx context.Context,
	console tools.Console,
	selector tools.Selector,
	logger tools.Logger,
	textToJSON tools.TextToJSONBackend,
	inputHandler tools.InputHandler,
	conversation chat.Conversation,
) error {
	logger.Info("Starting commit usecase")

	if err := checkGitRepository(ctx, console); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	logger.Info("Copying changes")
	err := stashChanges(ctx, console)
	if err != nil {
		return fmt.Errorf("failed to stash changes: %w", err)
	}

	err = unstashChanges(ctx, console)
	if err != nil {
		return fmt.Errorf("failed to unstash changes: %w", err)
	}

	defer func() {
		err = deleteStash(ctx, console)
		if err != nil {
			logger.Error("Failed to delete stash: " + err.Error())
		}
	}()

	logger.Info("Getting stash diff")
	buffer, err := getStashDiff(ctx, console)
	if err != nil {
		return fmt.Errorf("failed to get stash diff: %w", err)
	}

	conversation.AddMessage(
		chat.NewMessage(
			chat.RoleSystem,
			agentFilePrompt,
		),
	)
	conversation.AddMessage(
		chat.NewMessage(
			chat.RoleUser,
			buffer,
		),
	)

	if buffer == "" {
		logger.Info("No changes to commit")
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			logger.Info("Creating commit plan")
			resp, err := textToJSON.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("failed to convert text to JSON: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.RoleAssistant,
					resp,
				),
			)
			logger.Debug("Raw Commit plan: " + resp)

			var plan fileCommitPlan
			if err := json.Unmarshal([]byte(resp), &plan); err != nil {
				return fmt.Errorf(
					"failed to unmarshal commit plan: %w",
					err,
				)
			}

			logger.Println("Commit Plan 📝")
			logger.Println("------------")
			for _, a := range plan.CommitPlan {
				logger.Println(a.CommitMessage)
				for _, f := range a.Files {
					logger.Println("  - " + f)
				}
			}

			if duplicatedFiles := fileDuplicates(plan.CommitPlan); len(
				duplicatedFiles,
			) > 0 {
				logger.Info("Duplicate files found in commit plan")
				conversation.AddMessage(
					chat.NewMessage(
						chat.RoleSystem,
						"Duplicate files found in commit plan: "+strings.Join(
							duplicatedFiles,
							", ",
						),
					),
				)
				continue
			}

			if !selector.SelectBool(
				"Do you want to commit these changes?",
				true,
			) {
				newInstructions, err := inputHandler.Read(ctx, ">>> ")
				if err != nil {
					return fmt.Errorf(
						"failed to read new instructions: %w",
						err,
					)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.RoleUser,
						newInstructions,
					),
				)

				continue
			}

			for i, a := range plan.CommitPlan {
				cmd := tools.NewCommand(
					"git",
					"add",
				).WithArgs(a.Files...)

				result, err := console.Exec(ctx, cmd)
				if err != nil {
					return fmt.Errorf("failed to apply patch %d: %w", i, err)
				}
				if !result.Success() {
					return fmt.Errorf(
						"failed to apply patch %d: %s",
						i,
						result.Error,
					)
				}

				cmd = tools.NewCommand(
					"git",
					"commit",
					"--no-verify",
					"--message",
					a.CommitMessage,
				)
				result, err = console.Exec(ctx, cmd)
				if err != nil {
					return fmt.Errorf(
						"failed to commit changes %d: %w",
						i,
						err,
					)
				}
				if !result.Success() {
					return fmt.Errorf(
						"failed to commit changes %d: %s",
						i,
						result.Error,
					)
				}

				logger.Println("🚀 Committed " + a.CommitMessage)
			}

			return nil
		}
	}
}

func fileDuplicates(plan []fileAction) []string {
	var files []string
	seen := make(map[string]struct{})

	for _, p := range plan {
		for _, f := range p.Files {
			if _, ok := seen[f]; ok {
				files = append(files, f)
			}
			seen[f] = struct{}{}
		}
	}

	return files
}
