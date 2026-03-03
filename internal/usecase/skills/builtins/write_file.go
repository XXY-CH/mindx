package builtins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteFile(params map[string]any) (string, error) {
	filename, ok := params["filename"].(string)
	if !ok {
		return "", fmt.Errorf("invalid param: filename")
	}

	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid param: content")
	}

	startTime := time.Now()

	workDir := os.Getenv("MINDX_WORKSPACE")
	if workDir == "" {
		return "", fmt.Errorf("MINDX_WORKSPACE environment variable is not set")
	}

	baseDir := filepath.Join(workDir, "documents")

	// Sanitize filename to prevent path traversal
	cleanFilename := filepath.Clean(filename)
	if strings.HasPrefix(cleanFilename, "..") || filepath.IsAbs(cleanFilename) {
		return "", fmt.Errorf("invalid filename: path traversal detected")
	}

	var filePath string
	if path, ok := params["path"].(string); ok && path != "" {
		// Validate user-provided path against base directory
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return "", fmt.Errorf("invalid path: path traversal detected")
		}
		fullPath := filepath.Join(baseDir, cleanPath, cleanFilename)
		// Verify the resolved path is still under baseDir
		absBase, _ := filepath.Abs(baseDir)
		absFull, _ := filepath.Abs(fullPath)
		rel, err := filepath.Rel(absBase, absFull)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("invalid path: path traversal detected")
		}
		filePath = fullPath
	} else {
		// Validate filename even without path parameter
		fullPath := filepath.Join(baseDir, cleanFilename)
		absBase, _ := filepath.Abs(baseDir)
		absFull, _ := filepath.Abs(fullPath)
		rel, err := filepath.Rel(absBase, absFull)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("invalid filename: path traversal detected")
		}
		filePath = fullPath
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create dir %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	elapsed := time.Since(startTime)

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	return getJSONWriteResult(absPath, len(content), elapsed)
}

func getJSONWriteResult(filePath string, contentLength int, elapsed time.Duration) (string, error) {
	output := map[string]interface{}{
		"file_path":      filePath,
		"content_length": contentLength,
		"elapsed_ms":     elapsed.Milliseconds(),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json serialize failed: %w", err)
	}
	return string(data), nil
}
