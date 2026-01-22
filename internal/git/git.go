package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedFiles returns markdown files that have been modified.
// It compares against the specified base reference (branch, commit, or defaults to HEAD).
// If base is empty, it compares against HEAD (uncommitted changes).
// If base is "HEAD", it compares working directory against HEAD.
// If base is a branch name like "main", it compares current HEAD against that branch.
func ChangedFiles(base string, patterns []string) ([]string, error) {
	// Check if we're in a git repository
	if !IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	var changedFiles []string
	var err error

	if base == "" || base == "HEAD" {
		// Get uncommitted changes (staged + unstaged)
		changedFiles, err = getUncommittedChanges()
	} else {
		// Get changes between base and HEAD
		changedFiles, err = getChangesSinceBase(base)
	}

	if err != nil {
		return nil, err
	}

	// Filter to only include markdown files that match the patterns
	return filterMarkdownFiles(changedFiles, patterns)
}

// IsGitRepository checks if the current directory is inside a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch attempts to determine the default branch (main or master)
func GetDefaultBranch() string {
	// Try to get the default branch from remote
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(output))
		// Extract branch name from refs/remotes/origin/main
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fall back to checking if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		if cmd.Run() == nil {
			return branch
		}
	}

	// Default to main
	return "main"
}

// getUncommittedChanges returns files with uncommitted changes (staged + unstaged)
func getUncommittedChanges() ([]string, error) {
	// Get both staged and unstaged changes
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// HEAD might not exist (empty repo), try without HEAD
		cmd = exec.Command("git", "diff", "--name-only")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get uncommitted changes: %w", err)
		}
	}

	// Also get untracked files
	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedOutput, _ := untrackedCmd.Output()

	// Combine both outputs
	allOutput := append(output, untrackedOutput...)

	return parseFileList(allOutput), nil
}

// getChangesSinceBase returns files changed between base and HEAD
func getChangesSinceBase(base string) ([]string, error) {
	// Use merge-base to find common ancestor for proper comparison
	mergeBaseCmd := exec.Command("git", "merge-base", base, "HEAD")
	mergeBaseOutput, err := mergeBaseCmd.Output()

	var diffBase string
	if err != nil {
		// If merge-base fails, use base directly
		diffBase = base
	} else {
		diffBase = strings.TrimSpace(string(mergeBaseOutput))
	}

	// Get changed files between base and HEAD
	cmd := exec.Command("git", "diff", "--name-only", diffBase, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changes since %s: %w", base, err)
	}

	files := parseFileList(output)

	// Also include uncommitted changes on top of HEAD
	uncommitted, _ := getUncommittedChanges()
	files = append(files, uncommitted...)

	// Deduplicate
	return deduplicate(files), nil
}

// parseFileList parses git output into a slice of file paths
func parseFileList(output []byte) []string {
	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// filterMarkdownFiles filters files to only include markdown files matching patterns
func filterMarkdownFiles(files []string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, file := range files {
		// Skip non-markdown files
		if !strings.HasSuffix(strings.ToLower(file), ".md") {
			continue
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			matched, err := matchPattern(file, pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
			}
			if matched && !seen[file] {
				// Convert to absolute path
				absPath, err := filepath.Abs(file)
				if err != nil {
					continue
				}
				seen[file] = true
				result = append(result, absPath)
				break
			}
		}
	}

	return result, nil
}

// matchPattern checks if a file matches a glob pattern
func matchPattern(file, pattern string) (bool, error) {
	// Handle ** patterns by splitting into parts
	if strings.Contains(pattern, "**") {
		// Convert ** pattern to a simpler check
		// e.g., "docs/**/*.md" should match "docs/foo/bar.md"
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check prefix
			if prefix != "" && !strings.HasPrefix(file, prefix) {
				return false, nil
			}

			// Check suffix pattern
			if suffix != "" {
				remaining := file
				if prefix != "" {
					remaining = strings.TrimPrefix(file, prefix+"/")
				}
				return filepath.Match(suffix, filepath.Base(remaining))
			}

			return true, nil
		}
	}

	// Simple glob match
	return filepath.Match(pattern, file)
}

// deduplicate removes duplicate entries from a slice
func deduplicate(files []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	return result
}
