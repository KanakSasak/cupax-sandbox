package models

import (
	"time"

	"github.com/google/uuid"
)

// AnalysisStatus represents the current state of an analysis job
type AnalysisStatus string

const (
	StatusRunning   AnalysisStatus = "running"
	StatusCompleted AnalysisStatus = "completed"
	StatusError     AnalysisStatus = "error"
)

// Analysis represents a malware analysis job
type Analysis struct {
	ID             uuid.UUID       `json:"id"`
	Filename       string          `json:"filename"`
	FileHashSHA256 string          `json:"file_hash_sha256"`
	Status         AnalysisStatus  `json:"status"`
	SubmittedAt    time.Time       `json:"submitted_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	ReportJSON     *AnalysisReport `json:"report_json,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
}

// AnalysisReport contains the parsed results from Noriben
type AnalysisReport struct {
	Summary         SummaryStats      `json:"summary"`
	ProcessActivity []ProcessActivity `json:"process_activity"`
	FileSystem      []FileSystemEvent `json:"file_system"`
	Registry        []RegistryEvent   `json:"registry"`
	Network         []NetworkEvent    `json:"network"`
	UniqueHosts     []string          `json:"unique_hosts"`
}

// SummaryStats contains high-level statistics
type SummaryStats struct {
	ExecutionTime      float64 `json:"execution_time"`
	ProcessingTime     float64 `json:"processing_time"`
	AnalysisTime       float64 `json:"analysis_time"`
	ProcessesCreated   int     `json:"processes_created"`
	FilesCreated       int     `json:"files_created"`
	RegistryModified   int     `json:"registry_modified"`
	NetworkConnections int     `json:"network_connections"`
}

// ProcessActivity represents a process creation event
type ProcessActivity struct {
	Timestamp   string `json:"timestamp"`
	ProcessName string `json:"process_name"`
	PID         string `json:"pid"`
	CommandLine string `json:"command_line"`
	ChildPID    string `json:"child_pid,omitempty"`
}

// FileSystemEvent represents file operations
type FileSystemEvent struct {
	Timestamp   string `json:"timestamp"`
	Operation   string `json:"operation"` // CreateFile, DeleteFile, RenameFile
	ProcessName string `json:"process_name"`
	PID         string `json:"pid"`
	Path        string `json:"path"`
	Hash        string `json:"hash,omitempty"`
	HashType    string `json:"hash_type,omitempty"`
	YaraHits    string `json:"yara_hits,omitempty"`
	VTHits      string `json:"vt_hits,omitempty"`
	ToPath      string `json:"to_path,omitempty"` // For rename operations
}

// RegistryEvent represents registry modifications
type RegistryEvent struct {
	Timestamp   string `json:"timestamp"`
	Operation   string `json:"operation"` // RegCreateKey, RegSetValue, RegDeleteValue, RegDeleteKey
	ProcessName string `json:"process_name"`
	PID         string `json:"pid"`
	Path        string `json:"path"`
	Data        string `json:"data,omitempty"`
}

// NetworkEvent represents network activity
type NetworkEvent struct {
	Timestamp   string `json:"timestamp"`
	Protocol    string `json:"protocol"`  // TCP, UDP
	Direction   string `json:"direction"` // Send, Receive
	ProcessName string `json:"process_name"`
	PID         string `json:"pid"`
	RemoteAddr  string `json:"remote_addr"`
}

