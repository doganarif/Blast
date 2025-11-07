//go:build darwin

package ca

import (
	"bytes"
	"fmt"
	"os/exec"
)

// installCertificate installs the certificate into macOS keychain
func installCertificate(certPath string) error {
	// First, try to remove any existing BlastProxy CA
	removeCmd := exec.Command("security", "delete-certificate", "-c", "BlastProxy Root CA",
		"/Library/Keychains/System.keychain")
	removeCmd.Run() // Ignore errors if it doesn't exist

	// Add the certificate as trusted
	cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain", certPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install certificate: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}
