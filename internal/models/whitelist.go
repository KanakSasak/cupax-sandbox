package models

import (
	"time"

	"github.com/google/uuid"
)

// WhitelistType represents the type of whitelist entry
type WhitelistType string

const (
	WhitelistTypeProcess  WhitelistType = "process"
	WhitelistTypeDomain   WhitelistType = "domain"
	WhitelistTypeIP       WhitelistType = "ip"
	WhitelistTypeRegistry WhitelistType = "registry"
)

// Whitelist represents a whitelist entry
type Whitelist struct {
	ID          uuid.UUID     `json:"id"`
	Type        WhitelistType `json:"type"`
	Value       string        `json:"value"`
	Description string        `json:"description"`
	IsRegex     bool          `json:"is_regex"`
	Enabled     bool          `json:"enabled"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// WhitelistCreate represents the request to create a whitelist entry
type WhitelistCreate struct {
	Type        WhitelistType `json:"type" binding:"required"`
	Value       string        `json:"value" binding:"required"`
	Description string        `json:"description"`
	IsRegex     bool          `json:"is_regex"`
	Enabled     bool          `json:"enabled"`
}

// WhitelistUpdate represents the request to update a whitelist entry
type WhitelistUpdate struct {
	Value       *string `json:"value"`
	Description *string `json:"description"`
	IsRegex     *bool   `json:"is_regex"`
	Enabled     *bool   `json:"enabled"`
}
