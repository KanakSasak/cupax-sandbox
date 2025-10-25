package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cupax/cupax/internal/models"
	"github.com/google/uuid"
)

// Repository provides database operations for analyses
type Repository struct {
	db *Database
}

// NewRepository creates a new repository instance
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// CreateAnalysis creates a new analysis record
func (r *Repository) CreateAnalysis(analysis *models.Analysis) error {
	query := `
		INSERT INTO analyses (id, filename, file_hash_sha256, status, submitted_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.DB.Exec(
		query,
		analysis.ID.String(),
		analysis.Filename,
		analysis.FileHashSHA256,
		analysis.Status,
		analysis.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create analysis: %w", err)
	}

	return nil
}

// GetAnalysisByID retrieves an analysis by ID
func (r *Repository) GetAnalysisByID(id uuid.UUID) (*models.Analysis, error) {
	query := `
		SELECT id, filename, file_hash_sha256, status, submitted_at, completed_at, report_json, error_message
		FROM analyses
		WHERE id = ?
	`

	var analysis models.Analysis
	var idStr string
	var reportJSON sql.NullString
	var completedAt sql.NullTime
	var errorMessage sql.NullString

	err := r.db.DB.QueryRow(query, id.String()).Scan(
		&idStr,
		&analysis.Filename,
		&analysis.FileHashSHA256,
		&analysis.Status,
		&analysis.SubmittedAt,
		&completedAt,
		&reportJSON,
		&errorMessage,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("analysis not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	// Parse UUID
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UUID: %w", err)
	}
	analysis.ID = parsedID

	// Handle nullable fields
	if completedAt.Valid {
		analysis.CompletedAt = &completedAt.Time
	}

	if errorMessage.Valid {
		analysis.ErrorMessage = &errorMessage.String
	}

	// Unmarshal JSON report
	if reportJSON.Valid && reportJSON.String != "" {
		var report models.AnalysisReport
		if err := json.Unmarshal([]byte(reportJSON.String), &report); err != nil {
			return nil, fmt.Errorf("failed to unmarshal report JSON: %w", err)
		}
		analysis.ReportJSON = &report
	}

	return &analysis, nil
}

// GetAnalysisByHash retrieves an analysis by SHA256 hash
func (r *Repository) GetAnalysisByHash(hash string) (*models.Analysis, error) {
	query := `
		SELECT id, filename, file_hash_sha256, status, submitted_at, completed_at, report_json, error_message
		FROM analyses
		WHERE file_hash_sha256 = ?
	`

	var analysis models.Analysis
	var idStr string
	var reportJSON sql.NullString
	var completedAt sql.NullTime
	var errorMessage sql.NullString

	err := r.db.DB.QueryRow(query, hash).Scan(
		&idStr,
		&analysis.Filename,
		&analysis.FileHashSHA256,
		&analysis.Status,
		&analysis.SubmittedAt,
		&completedAt,
		&reportJSON,
		&errorMessage,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found (not an error for deduplication check)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis by hash: %w", err)
	}

	// Parse UUID
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UUID: %w", err)
	}
	analysis.ID = parsedID

	// Handle nullable fields
	if completedAt.Valid {
		analysis.CompletedAt = &completedAt.Time
	}

	if errorMessage.Valid {
		analysis.ErrorMessage = &errorMessage.String
	}

	// Unmarshal JSON report
	if reportJSON.Valid && reportJSON.String != "" {
		var report models.AnalysisReport
		if err := json.Unmarshal([]byte(reportJSON.String), &report); err != nil {
			return nil, fmt.Errorf("failed to unmarshal report JSON: %w", err)
		}
		analysis.ReportJSON = &report
	}

	return &analysis, nil
}

// GetAllAnalyses retrieves all analyses ordered by submission time
func (r *Repository) GetAllAnalyses() ([]models.Analysis, error) {
	query := `
		SELECT id, filename, file_hash_sha256, status, submitted_at, completed_at, report_json, error_message
		FROM analyses
		ORDER BY submitted_at DESC
	`

	rows, err := r.db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query analyses: %w", err)
	}
	defer rows.Close()

	var analyses []models.Analysis
	for rows.Next() {
		var analysis models.Analysis
		var idStr string
		var reportJSON sql.NullString
		var completedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&idStr,
			&analysis.Filename,
			&analysis.FileHashSHA256,
			&analysis.Status,
			&analysis.SubmittedAt,
			&completedAt,
			&reportJSON,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}

		// Parse UUID
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UUID: %w", err)
		}
		analysis.ID = parsedID

		// Handle nullable fields
		if completedAt.Valid {
			analysis.CompletedAt = &completedAt.Time
		}

		if errorMessage.Valid {
			analysis.ErrorMessage = &errorMessage.String
		}

		// Unmarshal JSON report
		if reportJSON.Valid && reportJSON.String != "" {
			var report models.AnalysisReport
			if err := json.Unmarshal([]byte(reportJSON.String), &report); err != nil {
				return nil, fmt.Errorf("failed to unmarshal report JSON: %w", err)
			}
			analysis.ReportJSON = &report
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// UpdateAnalysisStatus updates the status of an analysis
func (r *Repository) UpdateAnalysisStatus(id uuid.UUID, status models.AnalysisStatus, message *string) error {
	query := `
		UPDATE analyses
		SET status = ?
		WHERE id = ?
	`

	result, err := r.db.DB.Exec(query, status, id.String())
	if err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("analysis not found")
	}

	return nil
}

// UpdateAnalysisReport updates the report and marks analysis as completed
func (r *Repository) UpdateAnalysisReport(id uuid.UUID, report *models.AnalysisReport) error {
	// Marshal report to JSON string
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	query := `
		UPDATE analyses
		SET status = ?, completed_at = ?, report_json = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err = r.db.DB.Exec(query, models.StatusCompleted, now, string(reportJSON), id.String())
	if err != nil {
		return fmt.Errorf("failed to update analysis report: %w", err)
	}

	return nil
}

// UpdateAnalysisError updates the analysis with an error message
func (r *Repository) UpdateAnalysisError(id uuid.UUID, errorMsg string) error {
	query := `
		UPDATE analyses
		SET status = ?, completed_at = ?, error_message = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.DB.Exec(query, models.StatusError, now, errorMsg, id.String())
	if err != nil {
		return fmt.Errorf("failed to update analysis error: %w", err)
	}

	return nil
}
