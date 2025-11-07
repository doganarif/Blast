package ca

import (
	"fmt"
)

// InstallInTrustStore installs the CA certificate into the OS trust store
func (ca *CA) InstallInTrustStore() error {
	if err := installCertificate(ca.GetCertPath()); err != nil {
		return fmt.Errorf("failed to install certificate in trust store: %w", err)
	}
	fmt.Println("CA certificate installed in system trust store")
	fmt.Printf("\nFirefox users: Run 'blast ca-path' for Firefox setup instructions\n\n")
	return nil
}
