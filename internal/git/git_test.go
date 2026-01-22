package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupGitRepo creates a temporary git repository for testing
func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	cmd.Run()

	return tmpDir
}

// commitFile creates a file and commits it
func commitFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := exec.Command("git", "add", filename)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add "+filename)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestIsGitRepository(t *testing.T) {
	// Test in a git repo
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	if !IsGitRepository() {
		t.Error("expected IsGitRepository to return true in a git repo")
	}

	// Test in a non-git directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	if IsGitRepository() {
		t.Error("expected IsGitRepository to return false outside a git repo")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create initial commit (required for branch to exist)
	commitFile(t, repoDir, "README.md", "# Test")

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The default branch could be "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("unexpected branch name: %s", branch)
	}
}

func TestGetDefaultBranch(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create initial commit on main
	commitFile(t, repoDir, "README.md", "# Test")

	branch := GetDefaultBranch()

	// Should return main or master
	if branch != "main" && branch != "master" {
		t.Errorf("unexpected default branch: %s", branch)
	}
}

func TestChangedFiles_UncommittedChanges(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create initial commit
	commitFile(t, repoDir, "README.md", "# Initial")

	// Create uncommitted change
	os.WriteFile(filepath.Join(repoDir, "CHANGED.md"), []byte("# Changed"), 0644)

	files, err := ChangedFiles("HEAD", []string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 changed file, got %d", len(files))
	}

	if len(files) > 0 && filepath.Base(files[0]) != "CHANGED.md" {
		t.Errorf("expected CHANGED.md, got %s", filepath.Base(files[0]))
	}
}

func TestChangedFiles_ModifiedFile(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create and commit initial file
	commitFile(t, repoDir, "README.md", "# Initial")

	// Modify the file
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Modified"), 0644)

	files, err := ChangedFiles("HEAD", []string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 changed file, got %d", len(files))
	}

	if len(files) > 0 && filepath.Base(files[0]) != "README.md" {
		t.Errorf("expected README.md, got %s", filepath.Base(files[0]))
	}
}

func TestChangedFiles_PatternFiltering(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create initial commit
	commitFile(t, repoDir, "README.md", "# Initial")

	// Create both .md and .txt files
	os.WriteFile(filepath.Join(repoDir, "doc.md"), []byte("# Doc"), 0644)
	os.WriteFile(filepath.Join(repoDir, "notes.txt"), []byte("Notes"), 0644)

	// Only match .md files
	files, err := ChangedFiles("HEAD", []string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only include doc.md
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	for _, f := range files {
		if filepath.Ext(f) != ".md" {
			t.Errorf("expected only .md files, got %s", f)
		}
	}
}

func TestChangedFiles_NoChanges(t *testing.T) {
	repoDir := setupGitRepo(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(repoDir)

	// Create and commit a file
	commitFile(t, repoDir, "README.md", "# Test")

	// No uncommitted changes
	files, err := ChangedFiles("HEAD", []string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 changed files, got %d", len(files))
	}
}

func TestChangedFiles_NotAGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	_, err := ChangedFiles("HEAD", []string{"*.md"})
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		file    string
		pattern string
		want    bool
	}{
		{"README.md", "*.md", true},
		{"README.md", "*.txt", false},
		{"docs/guide.md", "docs/*.md", true},
		{"docs/guide.md", "*.md", false}, // simple glob doesn't match subdirs
		{"README.md", "README.md", true},
		{"docs/api/ref.md", "docs/**/*.md", true},
		{"api/ref.md", "docs/**/*.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.file+"_"+tt.pattern, func(t *testing.T) {
			got, err := matchPattern(tt.file, tt.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.file, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := deduplicate(input)

	if len(result) != 4 {
		t.Errorf("expected 4 unique items, got %d", len(result))
	}

	// Check that all unique values are present
	seen := make(map[string]bool)
	for _, v := range result {
		if seen[v] {
			t.Errorf("duplicate found: %s", v)
		}
		seen[v] = true
	}
}

func TestParseFileList(t *testing.T) {
	output := []byte("file1.md\nfile2.md\n\nfile3.md\n")
	files := parseFileList(output)

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	expected := []string{"file1.md", "file2.md", "file3.md"}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], f)
		}
	}
}
