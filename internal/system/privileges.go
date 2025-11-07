package system

import (
	"fmt"
	"os"
	"runtime"
)

// RequirePrivileges checks if the process is running with elevated privileges
// and exits with an error message if not
func RequirePrivileges() {
	if !HasPrivileges() {
		fmt.Fprintln(os.Stderr, "Error: This command requires elevated privileges")
		if runtime.GOOS == "windows" {
			fmt.Fprintln(os.Stderr, "Please run as Administrator")
		} else {
			fmt.Fprintln(os.Stderr, "Please run with sudo")
		}
		os.Exit(1)
	}
}
