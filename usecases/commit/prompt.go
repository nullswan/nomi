package commit

const agentFilePrompt = `
Create a commit plan in JSON format for staging changes and creating commits using Git, returning an array of files, adhering to provided guidelines.

## Important Note

You won't be able to commit the same file twice.
But when a file have been moved, you will commit both the old and new file.

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
- Make sure to include at least one significant component and module in the scope.
- Keep messages concise, within 75 characters for titles.

## Commit Message Specifications
- **Tense:** Present
- **Prefixes:** 'feat:', 'fix:', 'docs:', 'style:', 'refactor:', 'perf:', 'test:', 'chore:', 'ci:'
- **Scope:** Specify affected module and component in parentheses, **including relevant directories or submodules (e.g., 'internal/term', 'usecases/commit', 'services/api')**.
- **No Body:** Keep it concise unless additional description is necessary.

### Additional Guidelines
- Group related changes for the same feature in a single commit.
- Use multiple commits for unrelated changes.
- Ensure messages are clear, concise, and specific.
- Maintain consistent scoping based on file paths, components and modules.

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
      "commitMessage": "docs(internal/prompts): add new templates for prompts"
    }
  ]
}

# Notes

- Ensure each action accurately reflects the set of files related to a specific change or feature.
- Make certain the commit messages follow all specified guidelines for clarity and conciseness.
`
