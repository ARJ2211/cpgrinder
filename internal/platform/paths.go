package platform

import (
	"errors"
	"os"
	"path/filepath"
)

/*
Helper function to get the home directory
*/
func getHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

/*
Helper function to create a path if it does not exist
*/
func makeIfNotExist(path string) error {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return errors.New("failed to create " + path)
	}

	return nil
}

/*
ResolvePaths takes in a dbOverride `string` if given and returns the
path of the sqlite database, the workspace directory, and an error.
*/
func ResolvePaths(dbOverride string) (string, string, error) {
	var dbPath, workspaceDir string

	HOME, err := getHomeDir()
	if err != nil {
		return "", "", errors.New("failed to fetch home directory")
	}

	baseDir := filepath.Join(HOME, ".cpgrinder")
	defaultDb := filepath.Join(baseDir, "cpgrinder.sqlite")
	workspaceDir = filepath.Join(baseDir, "workspace")

	if dbOverride != "" {
		_, err := os.Stat(dbOverride)
		if errors.Is(err, os.ErrNotExist) {
			return "", "", errors.New("db override path is invalid")
		}
		dbPath = dbOverride
	} else {
		dbPath = defaultDb
	}

	// Make .cpgrinder directory if not exists
	err = makeIfNotExist(baseDir)
	if err != nil {
		return "", "", err
	}

	// Make .cpgrinder/workspace directory if not exists
	err = makeIfNotExist(workspaceDir)
	if err != nil {
		return "", "", err
	}

	return dbPath, workspaceDir, nil
}
