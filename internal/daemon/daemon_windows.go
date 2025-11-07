//go:build windows

package daemon

import (
	"os/exec"
)

// setProcAttributes sets Windows-specific process attributes
func setProcAttributes(cmd *exec.Cmd) {
	// Windows doesn't need special process group handling
	// The process will run detached by default
}
