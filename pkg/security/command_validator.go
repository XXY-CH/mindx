package security

import (
	"fmt"
	"strings"
)

// DangerousCommands is the list of commands that require explicit dangerous=true.
var DangerousCommands = []string{
	"rm", "dd", "mkfs", "format", "shutdown", "reboot", "init",
	"kill", "killall", "pkill", "fdisk", "parted", "chmod", "chown",
}

// DangerousCharacters are shell metacharacters that indicate injection attempts.
var DangerousCharacters = []string{";", "&", "|", "`", "$", "(", ")"}

// ValidateCommand checks a command string for injection patterns and dangerous commands.
// Returns an error if the command is unsafe.
func ValidateCommand(command string, dangerousAllowed bool) error {
	if command == "" {
		return fmt.Errorf("command is empty")
	}

	// Check for dangerous shell metacharacters
	for _, char := range DangerousCharacters {
		if strings.Contains(command, char) {
			return fmt.Errorf("command contains dangerous character: %s", char)
		}
	}

	// Extract the base command (first word)
	baseCmd := strings.Fields(command)[0]

	// Check against known dangerous commands
	if !dangerousAllowed && IsDangerousCommand(baseCmd) {
		return fmt.Errorf("dangerous command '%s' requires dangerous=true parameter", baseCmd)
	}

	return nil
}

// IsDangerousCommand checks if a command is in the dangerous commands list.
func IsDangerousCommand(cmd string) bool {
	for _, dangerous := range DangerousCommands {
		if cmd == dangerous {
			return true
		}
	}
	return false
}
