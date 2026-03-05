package builtins

import (
	"os"
	"path/filepath"
	"strings"

	"mindx/internal/config"
)

type fileAccessPolicy struct {
	enabled      bool
	workspace    string
	allowedPaths []string
}

func loadFileAccessPolicy(workspace string) (fileAccessPolicy, error) {
	policy := fileAccessPolicy{
		enabled:      false,
		workspace:    workspace,
		allowedPaths: nil,
	}

	cfg, err := config.LoadServerConfig()
	if err != nil {
		// Fail open for backward compatibility: if config cannot be loaded, keep unrestricted mode.
		return policy, nil
	}

	policy.enabled = cfg.FileAccess.Enabled
	if !policy.enabled {
		return policy, nil
	}

	normalizedAllowed := make([]string, 0, len(cfg.FileAccess.AllowedPaths))
	for _, p := range cfg.FileAccess.AllowedPaths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(workspace, p)
		}
		p = filepath.Clean(p)
		absPath, absErr := filepath.Abs(p)
		if absErr != nil {
			continue
		}
		normalizedAllowed = append(normalizedAllowed, absPath)
	}
	policy.allowedPaths = normalizedAllowed
	return policy, nil
}

func (p fileAccessPolicy) isAllowed(targetPath string) bool {
	if !p.enabled {
		return true
	}

	cleanTarget := filepath.Clean(targetPath)
	if p.workspace != "" && isPathWithinWorkspace(p.workspace, cleanTarget) {
		return true
	}

	for _, allowed := range p.allowedPaths {
		if matchesAllowedPath(allowed, cleanTarget) {
			return true
		}
	}

	return false
}

func matchesAllowedPath(allowedPath, targetPath string) bool {
	cleanAllowed := filepath.Clean(allowedPath)
	cleanTarget := filepath.Clean(targetPath)

	if cleanTarget == cleanAllowed {
		return true
	}

	info, err := os.Stat(cleanAllowed)
	if err == nil && info.IsDir() {
		return strings.HasPrefix(cleanTarget, cleanAllowed+string(filepath.Separator))
	}

	return false
}
