package git

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-i2p/logger"
)

var log = logger.GetGoI2PLogger()

// AddChanges stages files matching the given pattern in the git repository.
// If the repository doesn't exist or git operations fail, it logs and returns nil.
// This makes git functionality optional and gracefully degrades.
func AddChanges(repoPath, pattern string) error {
	// Check if .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Println("No git repository found at", repoPath, "- skipping git operations")
		return nil
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Println("Failed to open git repository:", err)
		return nil // Return nil to not break the workflow
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		log.Println("Failed to get worktree:", err)
		return nil
	}

	// Add the pattern to staging
	if err := worktree.AddWithOptions(&git.AddOptions{
		Path: pattern,
	}); err != nil {
		log.Println("Failed to add files to git:", err)
		return nil
	}

	log.Println("Successfully staged", pattern, "in git")

	// Commit the staged changes
	if err := CommitChanges(repoPath, "Automated stats update"); err != nil {
		log.Println("Failed to commit changes:", err)
		return nil
	}

	return nil
}

// CommitChanges creates a commit with the staged files in the git repository.
// If the repository doesn't exist or git operations fail, it logs and returns nil.
func CommitChanges(repoPath, message string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Println("Failed to open git repository:", err)
		return nil
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		log.Println("Failed to get worktree:", err)
		return nil
	}

	// Commit the staged changes
	commit, err := worktree.Commit(message, &git.CommitOptions{})
	if err != nil {
		log.Println("Failed to commit:", err)
		return nil
	}

	log.Println("Successfully committed changes:", commit.String())
	return nil
}
