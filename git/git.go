package git

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

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
	return nil
}
