//go:build linux

package ca

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installCertificate installs the certificate into Linux trust store
func installCertificate(certPath string) error {
	// Try multiple common locations
	certDirs := []string{
		"/usr/local/share/ca-certificates",
		"/etc/pki/ca-trust/source/anchors",
	}

	var targetDir string
	for _, dir := range certDirs {
		if _, err := os.Stat(dir); err == nil {
			targetDir = dir
			break
		}
	}

	if targetDir == "" {
		return fmt.Errorf("could not find certificate directory")
	}

	// Copy certificate
	destPath := filepath.Join(targetDir, "blast-ca.crt")
	data, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return err
	}

	// Update certificate store
	if targetDir == "/usr/local/share/ca-certificates" {
		return exec.Command("update-ca-certificates").Run()
	}
	return exec.Command("update-ca-trust").Run()
}
