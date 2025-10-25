package proxmox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cupax/cupax/internal/config"
)

// Client represents a Proxmox API client
type Client struct {
	config     config.ProxmoxConfig
	httpClient *http.Client
	baseURL    string
	authTicket string // For username/password authentication
	csrfToken  string // CSRF prevention token
}

// SnapshotInfo represents a Proxmox snapshot
type SnapshotInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SnapTime    int64  `json:"snaptime"`
	Parent      string `json:"parent"`
}

// ProxmoxResponse represents a generic Proxmox API response
type ProxmoxResponse struct {
	Data interface{} `json:"data"`
}

// NewClient creates a new Proxmox API client
func NewClient(cfg config.ProxmoxConfig) *Client {
	// Create HTTP client with custom TLS config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifySSL,
		},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	client := &Client{
		config:     cfg,
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("%s/api2/json", strings.TrimSuffix(cfg.Host, "/")),
	}

	// If using username/password, obtain authentication ticket
	if cfg.Username != "" && cfg.Password != "" {
		if err := client.authenticate(); err != nil {
			// Log error but don't fail - let subsequent API calls fail with proper error
			fmt.Printf("Warning: Failed to authenticate with Proxmox: %v\n", err)
		}
	}

	return client
}

// authenticate obtains an authentication ticket using username/password
func (c *Client) authenticate() error {
	url := fmt.Sprintf("%s/access/ticket", c.baseURL)

	// Prepare form data
	data := fmt.Sprintf("username=%s&password=%s", c.config.Username, c.config.Password)
	body := strings.NewReader(data)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response to get ticket and CSRF token
	var authResp struct {
		Data struct {
			Ticket              string `json:"ticket"`
			CSRFPreventionToken string `json:"CSRFPreventionToken"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	c.authTicket = authResp.Data.Ticket
	c.csrfToken = authResp.Data.CSRFPreventionToken

	return nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication based on configured method
	if c.authTicket != "" {
		// Username/password authentication using ticket
		req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.authTicket))
		if method != "GET" {
			// Add CSRF token for state-changing operations
			req.Header.Set("CSRFPreventionToken", c.csrfToken)
		}
	} else if c.config.TokenID != "" && c.config.TokenSecret != "" {
		// API token authentication
		req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.config.TokenID, c.config.TokenSecret))
	} else {
		return nil, fmt.Errorf("no authentication configured (need either token or username/password)")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// GetSnapshots returns all snapshots for the configured VM
func (c *Client) GetSnapshots() ([]SnapshotInfo, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", c.config.Node, c.config.VMID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ProxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse data array
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var snapshots []SnapshotInfo
	if err := json.Unmarshal(dataBytes, &snapshots); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshots: %w", err)
	}

	return snapshots, nil
}

// GetLatestSnapshot returns the most recent snapshot (excluding 'current' state)
func (c *Client) GetLatestSnapshot() (*SnapshotInfo, error) {
	snapshots, err := c.GetSnapshots()
	if err != nil {
		return nil, err
	}

	// Filter out 'current' pseudo-snapshot
	var realSnapshots []SnapshotInfo
	for _, snap := range snapshots {
		if snap.Name != "current" {
			realSnapshots = append(realSnapshots, snap)
		}
	}

	if len(realSnapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found for VM %d", c.config.VMID)
	}

	// Sort by timestamp (newest first)
	sort.Slice(realSnapshots, func(i, j int) bool {
		return realSnapshots[i].SnapTime > realSnapshots[j].SnapTime
	})

	return &realSnapshots[0], nil
}

// RollbackToSnapshot restores VM to a specific snapshot
func (c *Client) RollbackToSnapshot(snapshotName string) error {
	path := "/nodes/ludus/qemu/111/snapshot/cupax/rollback"

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Proxmox returns 200 OK for successful operations
	// The response body contains a task UPID (can be empty or plain text)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rollback failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Success - don't try to parse response body
	return nil
}

// RollbackToLatest restores VM to the most recent snapshot
func (c *Client) RollbackToLatest() error {
	latest, err := c.GetLatestSnapshot()
	if err != nil {
		return err
	}

	return c.RollbackToSnapshot(latest.Name)
}

// ShutdownVM initiates a graceful shutdown of the VM
func (c *Client) ShutdownVM() error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/shutdown", c.config.Node, c.config.VMID)

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Proxmox returns 200 OK for successful operations
	// The response body contains a task UPID (can be empty or plain text)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shutdown failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Success - don't try to parse response body
	return nil
}

// StopVM forces the VM to stop (like pulling the power)
func (c *Client) StopVM() error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", c.config.Node, c.config.VMID)

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Proxmox returns 200 OK for successful operations
	// The response body contains a task UPID (can be empty or plain text)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stop failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Success - don't try to parse response body
	return nil
}

// StartVM starts the VM
func (c *Client) StartVM() error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", c.config.Node, c.config.VMID)

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Proxmox returns 200 OK for successful operations
	// The response body contains a task UPID (can be empty or plain text)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("start failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Success - don't try to parse response body
	return nil
}

// GetVMStatus returns the current VM status
func (c *Client) GetVMStatus() (string, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", c.config.Node, c.config.VMID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ProxmoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract status from data
	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected data format")
	}

	status, ok := dataMap["status"].(string)
	if !ok {
		return "", fmt.Errorf("status field not found")
	}

	return status, nil
}

// WaitForStatus waits for VM to reach a specific status
func (c *Client) WaitForStatus(targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := c.GetVMStatus()
		if err != nil {
			return err
		}

		if status == targetStatus {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for VM to reach status: %s", targetStatus)
}
