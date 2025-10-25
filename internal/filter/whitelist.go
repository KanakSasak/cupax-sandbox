package filter

import (
	"log"
	"regexp"
	"strings"

	"github.com/cupax/cupax/internal/database"
	"github.com/cupax/cupax/internal/models"
)

// WhitelistFilter handles filtering of analysis reports based on whitelists
type WhitelistFilter struct {
	repo       *database.Repository
	whitelists map[models.WhitelistType][]models.Whitelist
}

// NewWhitelistFilter creates a new whitelist filter
func NewWhitelistFilter(repo *database.Repository) *WhitelistFilter {
	return &WhitelistFilter{
		repo:       repo,
		whitelists: make(map[models.WhitelistType][]models.Whitelist),
	}
}

// LoadWhitelists loads enabled whitelists from database
func (f *WhitelistFilter) LoadWhitelists() error {
	whitelists, err := f.repo.GetEnabledWhitelists()
	if err != nil {
		return err
	}

	// Organize by type
	f.whitelists = make(map[models.WhitelistType][]models.Whitelist)
	for _, wl := range whitelists {
		f.whitelists[wl.Type] = append(f.whitelists[wl.Type], wl)
	}

	log.Printf("Loaded %d enabled whitelist entries", len(whitelists))
	return nil
}

// isWhitelisted checks if a value matches any whitelist entry of a given type
func (f *WhitelistFilter) isWhitelisted(value string, whitelistType models.WhitelistType) bool {
	if value == "" {
		return false
	}

	entries, exists := f.whitelists[whitelistType]
	if !exists {
		return false
	}

	for _, entry := range entries {
		if entry.IsRegex {
			// Use regex matching
			matched, err := regexp.MatchString(entry.Value, value)
			if err != nil {
				log.Printf("Invalid regex pattern %s: %v", entry.Value, err)
				continue
			}
			if matched {
				return true
			}
		} else {
			// Case-insensitive substring matching
			if strings.Contains(strings.ToLower(value), strings.ToLower(entry.Value)) {
				return true
			}
		}
	}

	return false
}

// FilterReport filters an analysis report based on whitelists
func (f *WhitelistFilter) FilterReport(report *models.AnalysisReport) *models.AnalysisReport {
	if report == nil {
		return report
	}

	// Reload whitelists to get latest changes
	if err := f.LoadWhitelists(); err != nil {
		log.Printf("Warning: Failed to reload whitelists: %v", err)
	}

	// Count before filtering
	beforeCounts := map[string]int{
		"processes": len(report.ProcessActivity),
		"files":     len(report.FileSystem),
		"registry":  len(report.Registry),
		"network":   len(report.Network),
		"hosts":     len(report.UniqueHosts),
	}

	// Filter each type of event
	report.ProcessActivity = f.filterProcessActivity(report.ProcessActivity)
	report.FileSystem = f.filterFileSystem(report.FileSystem)
	report.Registry = f.filterRegistry(report.Registry)
	report.Network = f.filterNetwork(report.Network)
	report.UniqueHosts = f.filterUniqueHosts(report.UniqueHosts)

	// Count after filtering
	afterCounts := map[string]int{
		"processes": len(report.ProcessActivity),
		"files":     len(report.FileSystem),
		"registry":  len(report.Registry),
		"network":   len(report.Network),
		"hosts":     len(report.UniqueHosts),
	}

	// Log filtering statistics
	log.Printf("Whitelist filtering applied: "+
		"Processes %d->%d, Files %d->%d, Registry %d->%d, Network %d->%d, Hosts %d->%d",
		beforeCounts["processes"], afterCounts["processes"],
		beforeCounts["files"], afterCounts["files"],
		beforeCounts["registry"], afterCounts["registry"],
		beforeCounts["network"], afterCounts["network"],
		beforeCounts["hosts"], afterCounts["hosts"])

	return report
}

func (f *WhitelistFilter) filterProcessActivity(events []models.ProcessActivity) []models.ProcessActivity {
	filtered := make([]models.ProcessActivity, 0, len(events))
	for _, event := range events {
		if !f.isWhitelisted(event.ProcessName, models.WhitelistTypeProcess) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (f *WhitelistFilter) filterFileSystem(events []models.FileSystemEvent) []models.FileSystemEvent {
	filtered := make([]models.FileSystemEvent, 0, len(events))
	for _, event := range events {
		// Filter by process name
		if !f.isWhitelisted(event.ProcessName, models.WhitelistTypeProcess) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (f *WhitelistFilter) filterRegistry(events []models.RegistryEvent) []models.RegistryEvent {
	filtered := make([]models.RegistryEvent, 0, len(events))
	for _, event := range events {
		// Filter by both process name and registry path
		if !f.isWhitelisted(event.ProcessName, models.WhitelistTypeProcess) &&
			!f.isWhitelisted(event.Path, models.WhitelistTypeRegistry) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (f *WhitelistFilter) filterNetwork(events []models.NetworkEvent) []models.NetworkEvent {
	filtered := make([]models.NetworkEvent, 0, len(events))
	for _, event := range events {
		// Extract host from remote_addr (format: host:port)
		host := event.RemoteAddr
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Check if host is whitelisted (as IP or domain)
		if !f.isWhitelisted(host, models.WhitelistTypeIP) &&
			!f.isWhitelisted(host, models.WhitelistTypeDomain) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (f *WhitelistFilter) filterUniqueHosts(hosts []string) []string {
	filtered := make([]string, 0, len(hosts))
	for _, host := range hosts {
		// Check if host is whitelisted (as IP or domain)
		if !f.isWhitelisted(host, models.WhitelistTypeIP) &&
			!f.isWhitelisted(host, models.WhitelistTypeDomain) {
			filtered = append(filtered, host)
		}
	}
	return filtered
}
