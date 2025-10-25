package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cupax/cupax/internal/models"
	"github.com/google/uuid"
)

// CreateWhitelist creates a new whitelist entry
func (r *Repository) CreateWhitelist(wl *models.Whitelist) error {
	query := `
		INSERT INTO whitelists (id, type, value, description, is_regex, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.DB.Exec(
		query,
		wl.ID.String(),
		wl.Type,
		wl.Value,
		wl.Description,
		wl.IsRegex,
		wl.Enabled,
		wl.CreatedAt,
		wl.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create whitelist: %w", err)
	}

	return nil
}

// GetWhitelistByID retrieves a whitelist entry by ID
func (r *Repository) GetWhitelistByID(id uuid.UUID) (*models.Whitelist, error) {
	query := `
		SELECT id, type, value, description, is_regex, enabled, created_at, updated_at
		FROM whitelists
		WHERE id = ?
	`

	var wl models.Whitelist
	var idStr string

	err := r.db.DB.QueryRow(query, id.String()).Scan(
		&idStr,
		&wl.Type,
		&wl.Value,
		&wl.Description,
		&wl.IsRegex,
		&wl.Enabled,
		&wl.CreatedAt,
		&wl.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("whitelist not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist: %w", err)
	}

	wl.ID, _ = uuid.Parse(idStr)

	return &wl, nil
}

// GetAllWhitelists retrieves all whitelist entries
func (r *Repository) GetAllWhitelists() ([]models.Whitelist, error) {
	query := `
		SELECT id, type, value, description, is_regex, enabled, created_at, updated_at
		FROM whitelists
		ORDER BY type, value
	`

	rows, err := r.db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelists: %w", err)
	}
	defer rows.Close()

	var whitelists []models.Whitelist
	for rows.Next() {
		var wl models.Whitelist
		var idStr string

		err := rows.Scan(
			&idStr,
			&wl.Type,
			&wl.Value,
			&wl.Description,
			&wl.IsRegex,
			&wl.Enabled,
			&wl.CreatedAt,
			&wl.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist: %w", err)
		}

		wl.ID, _ = uuid.Parse(idStr)
		whitelists = append(whitelists, wl)
	}

	return whitelists, nil
}

// GetWhitelistsByType retrieves whitelist entries by type
func (r *Repository) GetWhitelistsByType(whitelistType models.WhitelistType) ([]models.Whitelist, error) {
	query := `
		SELECT id, type, value, description, is_regex, enabled, created_at, updated_at
		FROM whitelists
		WHERE type = ?
		ORDER BY value
	`

	rows, err := r.db.DB.Query(query, whitelistType)
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelists by type: %w", err)
	}
	defer rows.Close()

	var whitelists []models.Whitelist
	for rows.Next() {
		var wl models.Whitelist
		var idStr string

		err := rows.Scan(
			&idStr,
			&wl.Type,
			&wl.Value,
			&wl.Description,
			&wl.IsRegex,
			&wl.Enabled,
			&wl.CreatedAt,
			&wl.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist: %w", err)
		}

		wl.ID, _ = uuid.Parse(idStr)
		whitelists = append(whitelists, wl)
	}

	return whitelists, nil
}

// GetEnabledWhitelists retrieves all enabled whitelist entries
func (r *Repository) GetEnabledWhitelists() ([]models.Whitelist, error) {
	query := `
		SELECT id, type, value, description, is_regex, enabled, created_at, updated_at
		FROM whitelists
		WHERE enabled = 1
		ORDER BY type, value
	`

	rows, err := r.db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled whitelists: %w", err)
	}
	defer rows.Close()

	var whitelists []models.Whitelist
	for rows.Next() {
		var wl models.Whitelist
		var idStr string

		err := rows.Scan(
			&idStr,
			&wl.Type,
			&wl.Value,
			&wl.Description,
			&wl.IsRegex,
			&wl.Enabled,
			&wl.CreatedAt,
			&wl.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist: %w", err)
		}

		wl.ID, _ = uuid.Parse(idStr)
		whitelists = append(whitelists, wl)
	}

	return whitelists, nil
}

// UpdateWhitelist updates a whitelist entry
func (r *Repository) UpdateWhitelist(id uuid.UUID, update *models.WhitelistUpdate) error {
	// Build dynamic update query
	query := "UPDATE whitelists SET updated_at = ?"
	args := []interface{}{time.Now()}

	if update.Value != nil {
		query += ", value = ?"
		args = append(args, *update.Value)
	}

	if update.Description != nil {
		query += ", description = ?"
		args = append(args, *update.Description)
	}

	if update.IsRegex != nil {
		query += ", is_regex = ?"
		args = append(args, *update.IsRegex)
	}

	if update.Enabled != nil {
		query += ", enabled = ?"
		args = append(args, *update.Enabled)
	}

	query += " WHERE id = ?"
	args = append(args, id.String())

	result, err := r.db.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update whitelist: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("whitelist not found")
	}

	return nil
}

// DeleteWhitelist deletes a whitelist entry
func (r *Repository) DeleteWhitelist(id uuid.UUID) error {
	query := "DELETE FROM whitelists WHERE id = ?"

	result, err := r.db.DB.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete whitelist: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("whitelist not found")
	}

	return nil
}
