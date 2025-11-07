package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	caKeyFile  = "blast-ca.key"
	caCertFile = "blast-ca.crt"
)

// CA represents a Certificate Authority
type CA struct {
	Cert    *x509.Certificate
	Key     *rsa.PrivateKey
	CertPEM []byte
	KeyPEM  []byte
	Path    string
}

// EnsureCA loads or generates a CA certificate
func EnsureCA() (*CA, error) {
	caDir, err := getCADir()
	if err != nil {
		return nil, err
	}

	certPath := filepath.Join(caDir, caCertFile)
	keyPath := filepath.Join(caDir, caKeyFile)

	// Check if CA already exists
	if fileExists(certPath) && fileExists(keyPath) {
		return loadCA(caDir)
	}

	// Generate new CA
	return generateCA(caDir)
}

// generateCA creates a new root CA
func generateCA(caDir string) (*CA, error) {
	// Ensure directory exists
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create CA directory: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create CA certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"BlastProxy"},
			CommonName:    "BlastProxy Root CA",
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Write to disk
	certPath := filepath.Join(caDir, caCertFile)
	keyPath := filepath.Join(caDir, caKeyFile)

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key: %w", err)
	}

	return &CA{
		Cert:    cert,
		Key:     privateKey,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		Path:    caDir,
	}, nil
}

// loadCA loads an existing CA from disk
func loadCA(caDir string) (*CA, error) {
	certPath := filepath.Join(caDir, caCertFile)
	keyPath := filepath.Join(caDir, caKeyFile)

	// Read certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Read private key
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ = pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CA{
		Cert:    cert,
		Key:     privateKey,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		Path:    caDir,
	}, nil
}

// GetCertPath returns the path to the CA certificate file
func (ca *CA) GetCertPath() string {
	return filepath.Join(ca.Path, caCertFile)
}

// getCADir returns the directory where CA files are stored
func getCADir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "blast", "ca"), nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
