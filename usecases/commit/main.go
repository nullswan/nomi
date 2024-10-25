package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

// TODO(nullswan): Handle stash reference correctly to avoid any TOCTOU issues.
// TODO(nullswan): Add memory on the commit plan, preference, commonly used prefix, scopes, modules, and components.
// TODO(nullswan): Handle progressive commit plan, reducing the delay, ask to merge using file name when too many lines are changed.

const agentCodePrompt = `
Create a commit plan in JSON format for staging patches and creating commits using Git, adhering to provided guidelines.

## JSON Structure

The commit plan should be represented as a JSON object containing a list of actions. Each action includes both a 'patch' and a 'commit':

- **Action**: Contains the Git patch, and the commit message.

### Steps

1. **Analyze the Git Diff:**
   - Group related changes into features or fixes.
   - Determine necessity for multiple commits for unrelated changes.

2. **Prepare Staging Commands:**
   - Use 'git apply --cached' to stage specific lines.
   - Ensure accurate staging for atomic, feature-specific commits.

3. **Generate Commit Messages:**
   - Maintain present tense with an appropriate prefix and scope.
   - Exclude meaningless component names (e.g., "internal") from commit titles.
   - Preserve only significant component names.
   - Keep messages concise, within 75 characters for titles.

## Commit Message Specifications

- **Tense:** Present
- **Prefixes:** 'feat:', 'fix:', 'docs:', 'style:', 'refactor:', 'perf:', 'test:', 'chore:', 'ci:'
- **Scope:** Specify affected significant component/module in parentheses.
- **No Body:** Keep it concise unless additional description is necessary.

### Additional Guidelines

- Group related changes for the same feature in a single commit.
- Use multiple commits for unrelated changes.
- Ensure messages are clear, concise, and specific.
- Maintain consistent scoping based on file paths and modules.

## Output Format

Provide the commit plan in a plain JSON format containing the necessary actions with both 'patch' and 'commit' details.

## Example Commit Plan

{
  "commitPlan": [
    {
      "patch": "diff --git a/cmd/cli/main.go b/cmd/cli/main.go\nindex 83c3e7f..b4b49b6 100644\n--- a/cmd/cli/main.go\n+++ b/cmd/cli/main.go\n@@ -10,6 +10,7 @@ package main\n import (\n     \"fmt\"\n     \"os\"\n+    \"time\"\n )\n",
      "commitMessage": "feat(cmd/cli): add time import"
    },
    {
      "patch": "diff --git a/internal/prompts/templates.go b/internal/prompts/templates.go\nindex e69de29..f8a7e5d 100644\n--- a/internal/prompts/templates.go\n+++ b/internal/prompts/templates.go\n@@ -0,0 +1 @@\n+// New templates for prompts\n",
      "commitMessage": "docs(prompts): add new templates for prompts"
    }
  ]
}
`

const agentFilePrompt = `
Create a commit plan in JSON format for staging changes and creating commits using Git, returning an array of files, adhering to provided guidelines.

## JSON Structure

The commit plan should be represented as a JSON object containing a list of actions. Each action includes both a list of modified 'files' and a 'commitMessage':

- **Action**: Contains the array of files changed and the commit message.

### Steps

1. **Analyze the Git Diff:**
   - Group related changes into features or fixes.
   - Determine necessity for multiple commits for unrelated changes.

2. **Prepare Staging Commands:**
   - Identify files affected by the diff for atomic, feature-specific commits.

3. **Generate Commit Messages:**
   - Maintain present tense with an appropriate prefix and scope.
   - Exclude meaningless component names (e.g., "internal") from commit titles.
   - Preserve only significant component names.
   - Keep messages concise, within 75 characters for titles.

## Commit Message Specifications

- **Tense:** Present
- **Prefixes:** 'feat:', 'fix:', 'docs:', 'style:', 'refactor:', 'perf:', 'test:', 'chore:', 'ci:'
- **Scope:** Specify affected significant component/module in parentheses.
- **No Body:** Keep it concise unless additional description is necessary.

### Additional Guidelines

- Group related changes for the same feature in a single commit.
- Use multiple commits for unrelated changes.
- Ensure messages are clear, concise, and specific.
- Maintain consistent scoping based on file paths and modules.

## Output Format

Provide the commit plan in a plain JSON format containing the necessary actions with both 'files' and 'commitMessage' details.

## Example Commit Plan

{
  "commitPlan": [
    {
      "files": ["cmd/cli/main.go"],
      "commitMessage": "feat(cmd/cli): add time import"
    },
    {
      "files": ["internal/prompts/templates.go"],
      "commitMessage": "docs(prompts): add new templates for prompts"
    }
  ]
}

# Notes

- Ensure each action accurately reflects the set of files related to a specific change or feature.
- Make certain the commit messages follow all specified guidelines for clarity and conciseness.
`

type codeCommitPlan struct {
	CommitPlan []codeAction `json:"commitPlan"`
}

type codeAction struct {
	Patch         string `json:"patch"`
	CommitMessage string `json:"commitMessage"`
}

type fileCommitPlan struct {
	CommitPlan []fileAction `json:"commitPlan"`
}

type fileAction struct {
	Files         []string `json:"files"`
	CommitMessage string   `json:"commitMessage"`
}

func OnStart(
	ctx context.Context,
	console tools.Console,
	selector tools.Selector,
	logger tools.Logger,
	textToJSON tools.TextToJSONBackend,
	inputArea tools.InputArea,
	conversation chat.Conversation,
) error {
	logger.Info("Starting commit usecase")

	if err := checkGitRepository(ctx, console); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	logger.Info("Stashing changes")
	err := stashChanges(ctx, console)
	if err != nil {
		return fmt.Errorf("failed to stash changes: %w", err)
	}

	// Unstash changes directly so you can continue working on the changes
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

	fileMode := false
	bufferLines := strings.Count(buffer, "\n")
	if bufferLines > 100 {
		logger.Info("Detected large diff")
		fileMode = selector.SelectBool(
			"Would you like to commit changes by file instead of per-line?",
			true,
		)
	}

	if fileMode {
		conversation.AddMessage(
			chat.NewMessage(chat.Role(chat.RoleSystem), agentFilePrompt),
		)
	} else {
		conversation.AddMessage(
			chat.NewMessage(chat.Role(chat.RoleSystem), agentCodePrompt),
		)
	}

	logger.Debug("Stash diff: " + buffer)
	if buffer == "" {
		logger.Info("No changes to commit")
		return nil
	}

	conversation.AddMessage(
		chat.NewMessage(chat.Role(chat.RoleUser), buffer),
	)

	if !fileMode {
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
				logger.Debug("Raw Commit plan: " + resp)

				var plan codeCommitPlan
				if err := json.Unmarshal([]byte(resp), &plan); err != nil {
					return fmt.Errorf(
						"failed to unmarshal commit plan: %w",
						err,
					)
				}

				logger.Println("Commit Plan:")
				for _, a := range plan.CommitPlan {
					logger.Println("\t" + a.CommitMessage)
				}

				if !selector.SelectBool(
					"Do you want to commit these changes?",
					true,
				) {
					newInstructions := inputArea.Read(">>> ")
					conversation.AddMessage(
						chat.NewMessage(
							chat.Role(chat.RoleUser),
							newInstructions,
						),
					)

					continue
				}

				var errors []error
				for i, a := range plan.CommitPlan {
					// Patch should end with a newline
					if !strings.HasSuffix(a.Patch, "\n") {
						a.Patch += "\n"
					}

					cmd := tools.NewCommand(
						"git",
						"apply",
						"--cached",
						"-p1",
						"-",
					).WithInput(a.Patch)

					result, err := console.Exec(ctx, cmd)
					if err != nil {
						errors = append(
							errors,
							fmt.Errorf("failed to apply patch %d: %w", i, err),
						)
						continue
					}
					if !result.Success() {
						errors = append(errors, fmt.Errorf(
							"failed to apply patch %d: %s",
							i,
							result.Error,
						))
						continue
					}

					cmd = tools.NewCommand(
						"git",
						"commit",
						"--message",
						a.CommitMessage,
					)
					result, err = console.Exec(ctx, cmd)
					if err != nil {
						errors = append(
							errors,
							fmt.Errorf(
								"failed to commit changes %d: %w",
								i,
								err,
							),
						)
						continue
					}
					if !result.Success() {
						errors = append(errors, fmt.Errorf(
							"failed to commit changes %d: %s",
							i,
							result.Error,
						))
						continue
					}

					logger.Info("Committed " + a.CommitMessage)
				}

				if len(errors) > 0 {
					var errStr string
					for _, e := range errors {
						errStr += e.Error() + "\n"
					}

					return fmt.Errorf(
						"failed to commit all changes: %s",
						errStr,
					)
				}

				return nil
			}
		}
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
			logger.Debug("Raw Commit plan: " + resp)

			var plan fileCommitPlan
			if err := json.Unmarshal([]byte(resp), &plan); err != nil {
				return fmt.Errorf(
					"failed to unmarshal commit plan: %w",
					err,
				)
			}

			logger.Println("Commit Plan:")
			for _, a := range plan.CommitPlan {
				logger.Println("\t" + a.CommitMessage)
			}

			if !selector.SelectBool(
				"Do you want to commit these changes?",
				true,
			) {
				newInstructions := inputArea.Read(">>> ")
				conversation.AddMessage(
					chat.NewMessage(
						chat.Role(chat.RoleUser),
						newInstructions,
					),
				)

				continue
			}

			var errors []error
			for i, a := range plan.CommitPlan {
				cmd := tools.NewCommand(
					"git",
					"add",
				).WithArgs(a.Files...)

				result, err := console.Exec(ctx, cmd)
				if err != nil {
					errors = append(
						errors,
						fmt.Errorf("failed to apply patch %d: %w", i, err),
					)
					continue
				}
				if !result.Success() {
					errors = append(errors, fmt.Errorf(
						"failed to apply patch %d: %s",
						i,
						result.Error,
					))
					continue
				}

				cmd = tools.NewCommand(
					"git",
					"commit",
					"--message",
					a.CommitMessage,
				)
				result, err = console.Exec(ctx, cmd)
				if err != nil {
					errors = append(
						errors,
						fmt.Errorf(
							"failed to commit changes %d: %w",
							i,
							err,
						),
					)
					continue
				}
				if !result.Success() {
					errors = append(errors, fmt.Errorf(
						"failed to commit changes %d: %s",
						i,
						result.Error,
					))
					continue
				}

				logger.Info("Committed " + a.CommitMessage)
			}

			if len(errors) > 0 {
				var errStr string
				for _, e := range errors {
					errStr += e.Error() + "\n"
				}

				return fmt.Errorf(
					"failed to commit all changes: %s",
					errStr,
				)
			}

			return nil
		}
	}
}

func checkGitRepository(ctx context.Context, console tools.Console) error {
	cmd := tools.NewCommand("git", "rev-parse", "--is-inside-work-tree")
	result, err := console.Exec(ctx, cmd)
	if err != nil || !result.Success() {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

func stashChanges(ctx context.Context, console tools.Console) error {
	timestamp := time.Now().Format("20060102T150405")
	stashName := fmt.Sprintf("nomi-stash-%s", timestamp)
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
		return fmt.Errorf("failed to stash changes  and received no output")
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
		return fmt.Errorf("unable to retrieve stash reference")
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
		return "", fmt.Errorf("failed to show stash diff")
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
		return fmt.Errorf("failed to unstash changes")
	}

	cmd = tools.NewCommand("git", "reset")
	result, err = console.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to reset changes: %w", err)
	}

	if !result.Success() {
		return fmt.Errorf("failed to reset changes")
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
		return fmt.Errorf("failed to delete stash")
	}
	return nil
}
