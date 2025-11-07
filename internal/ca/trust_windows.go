//go:build windows

package ca

import (
	"os/exec"
)

// installCertificate installs the certificate into Windows trust store
func installCertificate(certPath string) error {
	cmd := exec.Command("certutil", "-addstore", "-f", "ROOT", certPath)
	return cmd.Run()
}
