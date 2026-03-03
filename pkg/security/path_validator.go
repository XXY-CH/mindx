package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathTraversal = errors.New("path traversal detected")
	ErrAccessDenied  = errors.New("access denied: path not in allowed directories")
)

// ValidatePath validates that a user-provided path stays within the base directory.
// It returns the cleaned absolute path if valid.
func ValidatePath(baseDir, userPath string) (string, error) {
	if baseDir == "" {
		return "", errors.New("base directory is empty")
	}

	if userPath == "" {
		return "", errors.New("user path is empty")
	}

	// Check for obvious traversal patterns in raw input
	if strings.Contains(userPath, "\x00") {
		return "", fmt.Errorf("%w: null byte in path", ErrPathTraversal)
	}

	// Clean base directory
	cleanBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	// Join paths
	fullPath := filepath.Join(cleanBase, userPath)
	cleanFull := filepath.Clean(fullPath)

	// Ensure the result is still within base directory
	rel, err := filepath.Rel(cleanBase, cleanFull)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: result outside base directory", ErrPathTraversal)
	}

	// Check if path resolves to a symlink that might escape
	finalPath, err := filepath.EvalSymlinks(cleanFull)
	if err == nil {
		rel, err := filepath.Rel(cleanBase, finalPath)
		if err == nil && strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("%w: symlink escapes base directory", ErrPathTraversal)
		}
	}

	return cleanFull, nil
}

// ValidateFilePath validates a file path against allowed base directories.
func ValidateFilePath(baseDir, userPath string, allowedDirs []string) (string, error) {
	validatedPath, err := ValidatePath(baseDir, userPath)
	if err != nil {
		return "", err
	}

	if len(allowedDirs) == 0 {
		return validatedPath, nil
	}

	for _, allowedDir := range allowedDirs {
		cleanAllowed := filepath.Clean(allowedDir)
		if strings.HasPrefix(validatedPath, cleanAllowed) {
			return validatedPath, nil
		}
	}

	return "", ErrAccessDenied
}

// SanitizeFilename sanitizes a filename by removing dangerous characters.
func SanitizeFilename(filename string) string {
	// Remove directory separators
	filename = filepath.Base(filename)

	// Replace dangerous characters with underscore
	dangerousChars := []string{"..", "~", "\x00"}
	for _, char := range dangerousChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	return filename
}

// EnsureDirectoryExists ensures a directory exists, creating it if necessary.
func EnsureDirectoryExists(dir string, perm os.FileMode) error {
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}
