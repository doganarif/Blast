//go:build unix

package system

import (
	"os"
)

// HasPrivileges returns true if running as root
func HasPrivileges() bool {
	return os.Geteuid() == 0
}
