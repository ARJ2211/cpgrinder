package solve

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ARJ2211/cpgrinder/internal/store"
)

/*
Cleans the path string
*/
func cleanString(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "/", "_")

	return s
}

/*
Return the problem dir
*/
func ProblemDir(workspaceRoot string, p store.ProblemID) string {
	cleanSource := cleanString(p.Source)
	cleanSourceID := cleanString(p.SourceID)

	problemDir := filepath.Join(workspaceRoot, cleanSource, cleanSourceID)
	return problemDir
}

/*
Ensure the problem dir exists
*/
func EnsureProblemDir(workspaceRoot string, p store.ProblemID) (string, error) {
	cleanSource := cleanString(p.Source)
	cleanSourceID := cleanString(p.SourceID)

	targetPath := filepath.Join(workspaceRoot, cleanSource, cleanSourceID)

	_, err := os.Stat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			os.MkdirAll(targetPath, 0755)
		} else {
			return "", err
		}
	}

	return targetPath, nil
}
