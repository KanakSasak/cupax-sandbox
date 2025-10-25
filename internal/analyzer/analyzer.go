package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cupax/cupax/internal/filter"
	"github.com/cupax/cupax/internal/models"
	"github.com/cupax/cupax/internal/proxmox"
)

// Config holds analyzer configuration
type Config struct {
	AgentURL     string // Agent HTTP URL (e.g., http://agent-vm:9090)
	Timeout      int    // Analysis timeout in seconds
	AgentEnabled bool   // Whether to run analysis (false for testing)
}

// Analyzer executes malware analysis via remote agent
type Analyzer struct {
	config          Config
	httpClient      *http.Client
	whitelistFilter *filter.WhitelistFilter
	proxmoxClient   *proxmox.Client // Optional Proxmox client for VM control
}

// New creates a new analyzer instance
func New(cfg Config, whitelistFilter *filter.WhitelistFilter, proxmoxClient *proxmox.Client) *Analyzer {
	// Set defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 300 // 5 minutes default
	}
	if cfg.AgentURL == "" {
		cfg.AgentURL = "http://localhost:9090"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout+60) * time.Second,
	}

	return &Analyzer{
		config:          cfg,
		httpClient:      client,
		whitelistFilter: whitelistFilter,
		proxmoxClient:   proxmoxClient,
	}
}

// AnalyzeResult holds the result of an analysis
type AnalyzeResult struct {
	Report *models.AnalysisReport
	Error  error
}

// AgentResponse represents the response from the agent
type AgentResponse struct {
	Success bool                   `json:"success"`
	Report  *models.AnalysisReport `json:"report,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Analyze executes analysis by sending sample to remote agent
func (a *Analyzer) Analyze(samplePath string, analysisID string, isZip bool, zipPassword string) *AnalyzeResult {
	// If agent is disabled (e.g., for testing without Windows VM)
	if !a.config.AgentEnabled {
		return &AnalyzeResult{
			Report: &models.AnalysisReport{
				Summary: models.SummaryStats{
					ExecutionTime:      0,
					ProcessingTime:     0,
					AnalysisTime:       0,
					ProcessesCreated:   0,
					FilesCreated:       0,
					RegistryModified:   0,
					NetworkConnections: 0,
				},
				ProcessActivity: []models.ProcessActivity{},
				FileSystem:      []models.FileSystemEvent{},
				Registry:        []models.RegistryEvent{},
				Network:         []models.NetworkEvent{},
				UniqueHosts:     []string{},
			},
			Error: nil,
		}
	}

	// Proxmox VM control: Restore snapshot before analysis
	if a.proxmoxClient != nil {
		if err := a.prepareVM(); err != nil {
			log.Printf("Warning: Failed to prepare VM: %v", err)
			// Continue anyway - VM might already be in good state
		}
	}

	// Send sample to agent (synchronous - waits for complete analysis)
	report, err := a.sendToAgent(samplePath, analysisID, isZip, zipPassword)
	if err != nil {
		return &AnalyzeResult{
			Report: nil,
			Error:  err,
		}
	}

	// Apply whitelist filtering before returning
	if a.whitelistFilter != nil {
		report = a.whitelistFilter.FilterReport(report)
	}

	return &AnalyzeResult{
		Report: report,
		Error:  nil,
	}
}

// prepareVM prepares the VM for analysis by restoring to latest snapshot
func (a *Analyzer) prepareVM() error {
	log.Println("Proxmox: Preparing VM for analysis...")

	// Get latest snapshot
	latest, err := a.proxmoxClient.GetLatestSnapshot()
	if err != nil {
		return fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	log.Printf("Proxmox: Restoring to snapshot: %s (created: %s)",
		latest.Name, time.Unix(latest.SnapTime, 0).Format(time.RFC3339))

	// Rollback to latest snapshot
	if err := a.proxmoxClient.RollbackToSnapshot(latest.Name); err != nil {
		return fmt.Errorf("failed to rollback to snapshot: %w", err)
	}

	// Wait for VM to be in running state after rollback
	log.Println("Proxmox: Waiting for VM to start after rollback...")
	if err := a.proxmoxClient.WaitForStatus("running", 2*time.Minute); err != nil {
		return fmt.Errorf("VM did not reach running state: %w", err)
	}

	// Give agent time to start
	log.Println("Proxmox: Waiting for agent to be ready...")
	time.Sleep(30 * time.Second)

	// Check agent health
	if err := a.CheckAgentHealth(); err != nil {
		return fmt.Errorf("agent not ready after VM restore: %w", err)
	}

	log.Println("Proxmox: VM prepared successfully")
	return nil
}

// cleanupVM shuts down the VM after analysis
func (a *Analyzer) cleanupVM() error {
	log.Println("Proxmox: Shutting down VM after analysis...")

	// Try graceful shutdown first
	if err := a.proxmoxClient.ShutdownVM(); err != nil {
		log.Printf("Proxmox: Graceful shutdown failed: %v", err)
		log.Println("Proxmox: Forcing VM stop...")

		// Force stop if graceful fails
		if err := a.proxmoxClient.StopVM(); err != nil {
			return fmt.Errorf("failed to stop VM: %w", err)
		}
	}

	// Wait for VM to be stopped
	log.Println("Proxmox: Waiting for VM to stop...")
	if err := a.proxmoxClient.WaitForStatus("stopped", 2*time.Minute); err != nil {
		return fmt.Errorf("VM did not stop: %w", err)
	}

	log.Println("Proxmox: VM shutdown successfully")
	return nil
}

// sendToAgentAsync sends sample to agent (returns immediately)
func (a *Analyzer) sendToAgentAsync(samplePath string, analysisID string, isZip bool, zipPassword string) error {
	// Open sample file
	file, err := os.ReadFile(samplePath)
	if err != nil {
		return fmt.Errorf("failed to read sample file: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Extract filename from samplePath to preserve extension
	filename := filepath.Base(samplePath)

	// Add file with proper filename (preserves extension)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(file); err != nil {
		return fmt.Errorf("failed to write file to form: %w", err)
	}

	// Add analysis_id
	if err := writer.WriteField("analysis_id", analysisID); err != nil {
		return fmt.Errorf("failed to write analysis_id: %w", err)
	}

	// Add is_zip flag
	if isZip {
		if err := writer.WriteField("is_zip", "true"); err != nil {
			return fmt.Errorf("failed to write is_zip: %w", err)
		}

		// Add password if provided
		if zipPassword != "" {
			if err := writer.WriteField("password", zipPassword); err != nil {
				return fmt.Errorf("failed to write password: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Send request to agent (agent returns immediately)
	url := fmt.Sprintf("%s/analyze", a.config.AgentURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Use shorter timeout since agent returns immediately
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to agent: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read agent response: %w", err)
	}

	// Parse response (agent just confirms receipt)
	var agentResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &agentResp); err != nil {
		return fmt.Errorf("failed to parse agent response: %w", err)
	}

	if !agentResp.Success {
		return fmt.Errorf("agent rejected sample: %s", agentResp.Error)
	}

	log.Printf("Agent accepted sample, analysis running in background: %s", analysisID)
	return nil
}

// waitForAnalysisComplete waits for the agent to finish analysis and submit report
// The agent POSTs the report to /api/v1/internal/analyses/{id}/report
// We simply wait for the configured timeout duration
func (a *Analyzer) waitForAnalysisComplete(analysisID string) {
	timeout := time.Duration(a.config.Timeout) * time.Second
	log.Printf("Waiting for agent to complete analysis (timeout: %v)...", timeout)

	// Sleep for the timeout duration
	// The agent will submit the report during this time via HandleSubmitReport
	time.Sleep(timeout)

	log.Printf("Analysis wait period completed for: %s", analysisID)
}

// sendToAgent sends the sample to the remote agent for analysis (DEPRECATED - use sendToAgentAsync)
func (a *Analyzer) sendToAgent(samplePath string, analysisID string, isZip bool, zipPassword string) (*models.AnalysisReport, error) {
	// Open sample file
	file, err := os.ReadFile(samplePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sample file: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Extract filename from samplePath to preserve extension
	filename := filepath.Base(samplePath)

	// Add file with proper filename (preserves extension)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(file); err != nil {
		return nil, fmt.Errorf("failed to write file to form: %w", err)
	}

	// Add analysis_id
	if err := writer.WriteField("analysis_id", analysisID); err != nil {
		return nil, fmt.Errorf("failed to write analysis_id: %w", err)
	}

	// Add is_zip flag
	if isZip {
		if err := writer.WriteField("is_zip", "true"); err != nil {
			return nil, fmt.Errorf("failed to write is_zip: %w", err)
		}

		// Add password if provided
		if zipPassword != "" {
			if err := writer.WriteField("password", zipPassword); err != nil {
				return nil, fmt.Errorf("failed to write password: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Send request to agent
	url := fmt.Sprintf("%s/analyze", a.config.AgentURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to agent: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent response: %w", err)
	}

	// Parse response
	var agentResp AgentResponse
	if err := json.Unmarshal(respBody, &agentResp); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	if !agentResp.Success {
		return nil, fmt.Errorf("agent analysis failed: %s", agentResp.Error)
	}

	return agentResp.Report, nil
}

// CheckAgentHealth checks if the agent is reachable
func (a *Analyzer) CheckAgentHealth() error {
	url := fmt.Sprintf("%s/health", a.config.AgentURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("agent unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("agent returned status: %d", resp.StatusCode)
	}

	return nil
}
