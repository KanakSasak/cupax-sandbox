package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Data     DataConfig     `json:"data"`
	Analyzer AnalyzerConfig `json:"analyzer"`
	Proxmox  ProxmoxConfig  `json:"proxmox"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        string `json:"port"`
	FrontendDir string `json:"frontend_dir"`
}

// DataConfig holds data directory configuration
type DataConfig struct {
	BaseDir     string `json:"base_dir"`
	SamplesDir  string `json:"samples_dir"`
	FrontendDir string `json:"frontend_dir"`
}

// AnalyzerConfig holds analyzer configuration
type AnalyzerConfig struct {
	Enabled  bool   `json:"enabled"`   // Enable/disable agent communication
	AgentURL string `json:"agent_url"` // Agent HTTP URL (e.g., http://agent-vm:9090)
	Timeout  int    `json:"timeout"`   // Analysis timeout in seconds
}

// ProxmoxConfig holds Proxmox VE configuration
type ProxmoxConfig struct {
	Enabled        bool   `json:"enabled"`         // Enable/disable Proxmox integration
	Host           string `json:"host"`            // Proxmox host (e.g., https://proxmox.local:8006)
	Node           string `json:"node"`            // Proxmox node name
	VMID           int    `json:"vmid"`            // VM ID for agent

	// Authentication - use either token OR username/password
	TokenID        string `json:"token_id"`        // API token ID (e.g., root@pam!cupax)
	TokenSecret    string `json:"token_secret"`    // API token secret
	Username       string `json:"username"`        // Username (e.g., root@pam)
	Password       string `json:"password"`        // Password

	VerifySSL      bool   `json:"verify_ssl"`      // Verify SSL certificate
	RestoreSnapshot bool   `json:"restore_snapshot"` // Restore to latest snapshot before analysis
	ShutdownAfter  bool   `json:"shutdown_after"`  // Shutdown VM after analysis completes
}

// Load loads configuration from a JSON file
func Load(filepath string) (*Config, error) {
	// Set defaults
	cfg := &Config{
		Server: ServerConfig{
			Port:        "8080",
			FrontendDir: "./frontend/dist",
		},
		Data: DataConfig{
			BaseDir:    "./data",
			SamplesDir: "./data/samples",
		},
		Analyzer: AnalyzerConfig{
			Enabled:  false,
			AgentURL: "http://localhost:9090",
			Timeout:  300,
		},
		Proxmox: ProxmoxConfig{
			Enabled:         false,
			Host:            "https://proxmox.local:8006",
			Node:            "pve",
			VMID:            100,
			TokenID:         "",
			TokenSecret:     "",
			Username:        "",
			Password:        "",
			VerifySSL:       false,
			RestoreSnapshot: true,
			ShutdownAfter:   true,
		},
	}

	// Try to load config file
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a JSON file
func (c *Config) Save(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
