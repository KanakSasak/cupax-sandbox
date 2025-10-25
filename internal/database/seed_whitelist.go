package database

import (
	"log"
	"time"

	"github.com/cupax/cupax/internal/models"
	"github.com/google/uuid"
)

// SeedDefaultWhitelists adds common benign processes/domains/IPs to whitelist
func (r *Repository) SeedDefaultWhitelists() error {
	// Check if whitelists already exist
	existing, err := r.GetAllWhitelists()
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		log.Println("Whitelists already seeded, skipping...")
		return nil
	}

	log.Println("Seeding default whitelists...")

	defaultWhitelists := []models.Whitelist{
		// Windows System Processes
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "svchost.exe",
			Description: "Windows Service Host Process",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "explorer.exe",
			Description: "Windows Explorer",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "dwm.exe",
			Description: "Desktop Window Manager",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "csrss.exe",
			Description: "Client Server Runtime Process",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "lsass.exe",
			Description: "Local Security Authority Process",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "winlogon.exe",
			Description: "Windows Logon Process",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "services.exe",
			Description: "Service Control Manager",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "taskhostw.exe",
			Description: "Task Host Window",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeProcess,
			Value:       "RuntimeBroker.exe",
			Description: "Runtime Broker",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},

		// Registry Keys (Common Windows system keys)
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeRegistry,
			Value:       "HKLM\\\\SOFTWARE\\\\Microsoft\\\\Windows\\\\CurrentVersion\\\\Run",
			Description: "Windows startup registry key",
			IsRegex:     false,
			Enabled:     false, // Disabled by default - often interesting for malware
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},

		// Domains (Microsoft/Windows Update)
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeDomain,
			Value:       "microsoft.com",
			Description: "Microsoft official domain",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeDomain,
			Value:       "windows.com",
			Description: "Windows official domain",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeDomain,
			Value:       "windowsupdate.com",
			Description: "Windows Update domain",
			IsRegex:     false,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},

		// IP Ranges (Local/Private networks)
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeIP,
			Value:       "^127\\..*",
			Description: "Localhost addresses",
			IsRegex:     true,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeIP,
			Value:       "^192\\.168\\..*",
			Description: "Private network range 192.168.x.x",
			IsRegex:     true,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Type:        models.WhitelistTypeIP,
			Value:       "^10\\..*",
			Description: "Private network range 10.x.x.x",
			IsRegex:     true,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, wl := range defaultWhitelists {
		if err := r.CreateWhitelist(&wl); err != nil {
			log.Printf("Warning: Failed to seed whitelist entry %s: %v", wl.Value, err)
		}
	}

	log.Printf("Seeded %d default whitelist entries", len(defaultWhitelists))
	return nil
}
