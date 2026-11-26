package common_tests

import (
	"os"
	"path/filepath"
)

// FindMigrationFolderPath returns the absolute identity_directory_db migrations folder, whatever package runs the test.
func FindMigrationFolderPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "src", "identity_directory_db", "migrations"), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
