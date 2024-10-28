package commit

const agentFilePrompt = `
Your task is to analyze Git diff data and create a structured commit plan following specific commit message guidelines.

You will work through the following steps to accomplish this task:

# Steps

1. **Analyze the Git Diff:**
   - Group related changes into either features or fixes.
   - Decide if multiple commits are necessary for unrelated changes.

2. **Prepare Staging Commands:**
   - Identify the affected files for each atomic, feature-specific commit.
   - Ensure no file is committed more than once unless it has been moved.

3. **Generate Commit Messages:**
   - Use present tense with the appropriate prefixes such as 'feat:', 'fix:', 'docs:', 'style:', 'refactor:', 'perf:', 'test:', 'chore:', 'ci:'.
   - Include the scope by specifying relevant modules or directories.
   - Ensure each message is concise and under 75 characters.

4. **Assemble the Commit Plan:**
   - Present the commit plan as a JSON object, listing each action with its affected files and corresponding commit message.

5. **Review and Finalize:**
   - Verify that the commit plan adheres to all guidelines for clarity, consistency, and detail.
   - For moved files, include their old and new paths in separate commits.

# Output Format

- The output should be a JSON-formatted commit plan.
- Each commit action in the plan must include the list of affected files and the corresponding commit message.

# Examples

**Commit Plan Example:**
{
  "commitPlan": [
    {
      "files": ["src/components/Button.js", "src/styles/button.css"],
      "commitMessage": "feat(components): add Button component with styling"
    },
    {
      "files": ["docs/README.md"],
      "commitMessage": "docs: update README with setup instructions"
    },
    {
      "files": ["src/utils/helpers.js"],
      "commitMessage": "refactor(utils): improve helper functions"
    },
    {
      "files": ["src/oldPath/file.js"],
      "commitMessage": "chore(files): remove deprecated file location"
    },
    {
      "files": ["src/newPath/file.js"],
      "commitMessage": "chore(files): add file to new location"
    }
  ]
}

**Moved File Example:**
{
  "commitPlan": [
    {
      "files": ["src/oldPath/file.js"],
      "commitMessage": "chore(files): remove deprecated file location"
    },
    {
      "files": ["src/newPath/file.js"],
      "commitMessage": "chore(files): add file to new location"
    }
  ]
}

# Notes

- Ensure that each commit action is unique with respect to file changes.
- Moved files should appear as separate actions: once for removal and once for addition.
- Adhere to the commit message guidelines by using the correct prefixes and limiting the character count.
- You should never commit the same file more than once unless it has been moved to a new location.
`
