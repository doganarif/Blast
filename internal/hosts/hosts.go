package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const marker = "# blast-proxy"

// GetHostsPath returns the platform-specific hosts file path
func GetHostsPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

// AddEntry adds a domain entry to the hosts file
func AddEntry(domain string) error {
	hostsPath := GetHostsPath()

	// Read current content
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Check if entry already exists
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, domain) && strings.Contains(line, marker) {
			// Entry already exists
			return nil
		}
	}

	// Append new entry
	entry := fmt.Sprintf("127.0.0.1 %s %s\n", domain, marker)
	newContent := append(content, []byte(entry)...)

	// Write back
	if err := os.WriteFile(hostsPath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	return nil
}

// RemoveEntry removes a domain entry from the hosts file
func RemoveEntry(domain string) error {
	hostsPath := GetHostsPath()

	// Read current content
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Filter out the entry
	var newLines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		// Skip lines that contain both the domain and our marker
		if strings.Contains(line, domain) && strings.Contains(line, marker) {
			found = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !found {
		// Entry doesn't exist, nothing to do
		return nil
	}

	// Write back
	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(hostsPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	return nil
}
