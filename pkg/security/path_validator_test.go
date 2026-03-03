package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		baseDir   string
		userPath  string
		wantErr   bool
		errTarget error
	}{
		{
			name:     "valid simple path",
			baseDir:  tmpDir,
			userPath: "subdir/file.txt",
			wantErr:  false,
		},
		{
			name:      "path traversal with ..",
			baseDir:   tmpDir,
			userPath:  "../../../etc/passwd",
			wantErr:   true,
			errTarget: ErrPathTraversal,
		},
		{
			name:      "path traversal with encoded ..",
			baseDir:   tmpDir,
			userPath:  "subdir/../../..",
			wantErr:   true,
			errTarget: ErrPathTraversal,
		},
		{
			name:    "simple filename",
			baseDir: tmpDir,
			userPath: "test.txt",
			wantErr: false,
		},
		{
			name:    "nested valid path",
			baseDir: tmpDir,
			userPath: "a/b/c/file.txt",
			wantErr: false,
		},
		{
			name:      "null byte in path",
			baseDir:   tmpDir,
			userPath:  "file\x00.txt",
			wantErr:   true,
			errTarget: ErrPathTraversal,
		},
		{
			name:    "empty base dir",
			baseDir: "",
			userPath: "file.txt",
			wantErr: true,
		},
		{
			name:    "empty user path",
			baseDir: tmpDir,
			userPath: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePath(tt.baseDir, tt.userPath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePath() expected error, got nil (result: %s)", result)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePath() unexpected error: %v", err)
				}
				// Verify result is within base dir
				absBase, _ := filepath.Abs(tt.baseDir)
				rel, relErr := filepath.Rel(absBase, result)
				if relErr != nil {
					t.Errorf("result path not relative to base: %v", relErr)
				}
				if filepath.IsAbs(rel) || rel[:2] == ".." {
					t.Errorf("result path %s escapes base dir %s", result, absBase)
				}
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	allowed1 := filepath.Join(tmpDir, "allowed1")
	allowed2 := filepath.Join(tmpDir, "allowed2")
	os.MkdirAll(allowed1, 0755)
	os.MkdirAll(allowed2, 0755)

	tests := []struct {
		name        string
		baseDir     string
		userPath    string
		allowedDirs []string
		wantErr     bool
	}{
		{
			name:        "path in allowed dir",
			baseDir:     allowed1,
			userPath:    "file.txt",
			allowedDirs: []string{allowed1, allowed2},
			wantErr:     false,
		},
		{
			name:        "path not in allowed dir",
			baseDir:     tmpDir,
			userPath:    "file.txt",
			allowedDirs: []string{allowed1, allowed2},
			wantErr:     true,
		},
		{
			name:        "no allowed dirs specified",
			baseDir:     tmpDir,
			userPath:    "file.txt",
			allowedDirs: nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateFilePath(tt.baseDir, tt.userPath, tt.allowedDirs)
			if tt.wantErr && err == nil {
				t.Error("ValidateFilePath() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFilePath() unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal filename", "test.txt", "test.txt"},
		{"path traversal in filename", "../../../etc/passwd", "passwd"},
		{"null byte", "file\x00.txt", "file_.txt"},
		{"directory separators", "dir/subdir/file.txt", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
