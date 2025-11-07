package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ProxyMapping represents an active proxy configuration
type ProxyMapping struct {
	DomainPrefix string `json:"domain_prefix"`
	LocalPort    string `json:"local_port"`
	FullDomain   string `json:"full_domain"`
}

// Config represents the persistent configuration
type Config struct {
	CAPath       string                   `json:"ca_path"`
	Proxies      map[string]ProxyMapping  `json:"proxies"` // key: domain_prefix
	mu           sync.RWMutex
	path         string
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		path:    configPath,
		Proxies: make(map[string]ProxyMapping),
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// First run, create empty config
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}

// AddProxy adds a new proxy mapping
func (c *Config) AddProxy(prefix, port string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Proxies[prefix] = ProxyMapping{
		DomainPrefix: prefix,
		LocalPort:    port,
		FullDomain:   prefix + ".blast",
	}
}

// RemoveProxy removes a proxy mapping
func (c *Config) RemoveProxy(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Proxies, prefix)
}

// GetProxy retrieves a proxy mapping
func (c *Config) GetProxy(prefix string) (ProxyMapping, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	proxy, exists := c.Proxies[prefix]
	return proxy, exists
}

// ListProxies returns all active proxies
func (c *Config) ListProxies() []ProxyMapping {
	c.mu.RLock()
	defer c.mu.RUnlock()

	proxies := make([]ProxyMapping, 0, len(c.Proxies))
	for _, p := range c.Proxies {
		proxies = append(proxies, p)
	}
	return proxies
}

// SetCAPath sets the CA certificate path
func (c *Config) SetCAPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CAPath = path
}

// getConfigPath returns the platform-specific config file path
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use platform-specific config directory
	configDir := filepath.Join(homeDir, ".config", "blast")
	return filepath.Join(configDir, "config.json"), nil
}
