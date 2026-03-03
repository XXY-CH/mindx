package builtins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MINDX_WORKSPACE", tmpDir)

	// Ensure documents dir exists
	os.MkdirAll(filepath.Join(tmpDir, "documents"), 0755)

	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{
			name: "valid write",
			params: map[string]any{
				"filename": "test.txt",
				"content":  "hello",
			},
			wantErr: false,
		},
		{
			name: "valid write with path",
			params: map[string]any{
				"filename": "test.txt",
				"content":  "hello",
				"path":     "subdir",
			},
			wantErr: false,
		},
		{
			name: "path traversal in filename",
			params: map[string]any{
				"filename": "../../../etc/passwd",
				"content":  "malicious",
			},
			wantErr: true,
		},
		{
			name: "path traversal in path param",
			params: map[string]any{
				"filename": "test.txt",
				"content":  "malicious",
				"path":     "../../../etc",
			},
			wantErr: true,
		},
		{
			name: "absolute path in filename",
			params: map[string]any{
				"filename": "/etc/passwd",
				"content":  "malicious",
			},
			wantErr: true,
		},
		{
			name: "absolute path in path param",
			params: map[string]any{
				"filename": "test.txt",
				"content":  "malicious",
				"path":     "/etc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WriteFile(tt.params)
			if tt.wantErr && err == nil {
				t.Error("WriteFile() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("WriteFile() unexpected error: %v", err)
			}
		})
	}
}
