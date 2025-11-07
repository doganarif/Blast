//go:build windows

package system

import (
	"golang.org/x/sys/windows"
)

// HasPrivileges returns true if running as Administrator
func HasPrivileges() bool {
	var sid *windows.SID

	// Get the SID for the Administrators group
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// Check if the current token is a member of the Administrators group
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}
