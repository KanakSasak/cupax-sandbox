package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cupax/cupax/internal/analyzer"
	"github.com/cupax/cupax/internal/api"
	"github.com/cupax/cupax/internal/config"
	"github.com/cupax/cupax/internal/database"
	"github.com/cupax/cupax/internal/filter"
	"github.com/cupax/cupax/internal/proxmox"
)

const banner = `
   ____                  __  __
  / ___|_   _ _ __   __ _\ \/ /
 | |   | | | | '_ \ / _  |\  /
 | |___| |_| | |_) | (_| |/  \
  \____|\__,_| .__/ \__,_/_/\_\
             |_|
  Malware Analysis Platform
  Version 4.0.0 - Remote Agent Architecture
`

func main() {
	// Parse command line flags
	configFile := flag.String("config", "cupax.json", "Path to configuration file")
	genConfig := flag.Bool("gen-config", false, "Generate default configuration file and exit")
	flag.Parse()

	fmt.Println(banner)

	// Generate config file if requested
	if *genConfig {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Port:        "8080",
				FrontendDir: "./frontend/dist",
			},
			Data: config.DataConfig{
				BaseDir:    "./data",
				SamplesDir: "./data/samples",
			},
			Analyzer: config.AnalyzerConfig{
				Enabled:  false, // Set to true when agent is running
				AgentURL: "http://localhost:9090",
				Timeout:  300,
			},
			Proxmox: config.ProxmoxConfig{
				Enabled:         false, // Set to true to enable Proxmox integration
				Host:            "https://proxmox.local:8006",
				Node:            "pve",
				VMID:            100,
				TokenID:         "", // Use either token OR username/password
				TokenSecret:     "",
				Username:        "", // Alternative: username (e.g., root@pam)
				Password:        "", // Alternative: password
				VerifySSL:       false,
				RestoreSnapshot: true,
				ShutdownAfter:   true,
			},
		}

		if err := cfg.Save(*configFile); err != nil {
			log.Fatalf("Failed to generate config file: %v", err)
		}

		fmt.Printf("Generated config file: %s\n", *configFile)
		return
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Using configuration file: %s", *configFile)
	log.Printf("Data directory: %s", cfg.Data.BaseDir)
	log.Printf("Samples directory: %s", cfg.Data.SamplesDir)
	log.Printf("Agent enabled: %v", cfg.Analyzer.Enabled)
	if cfg.Analyzer.Enabled {
		log.Printf("Agent URL: %s", cfg.Analyzer.AgentURL)
	}

	// Ensure data directories exist
	if err := os.MkdirAll(cfg.Data.BaseDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	if err := os.MkdirAll(cfg.Data.SamplesDir, 0755); err != nil {
		log.Fatalf("Failed to create samples directory: %v", err)
	}

	// Initialize database
	dbConfig := database.Config{
		DataDir: cfg.Data.BaseDir,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	log.Println("Database schema initialized")

	// Create repository
	repo := database.NewRepository(db)

	// Seed default whitelists
	if err := repo.SeedDefaultWhitelists(); err != nil {
		log.Printf("Warning: Failed to seed default whitelists: %v", err)
	}

	// Create whitelist filter
	whitelistFilter := filter.NewWhitelistFilter(repo)
	if err := whitelistFilter.LoadWhitelists(); err != nil {
		log.Printf("Warning: Failed to load whitelists: %v", err)
	}

	// Create Proxmox client if enabled
	var proxmoxClient *proxmox.Client
	if cfg.Proxmox.Enabled {
		proxmoxClient = proxmox.NewClient(cfg.Proxmox)
		log.Printf("Proxmox integration enabled for VM %d on node %s", cfg.Proxmox.VMID, cfg.Proxmox.Node)
		log.Printf("Proxmox features: RestoreSnapshot=%v, ShutdownAfter=%v",
			cfg.Proxmox.RestoreSnapshot, cfg.Proxmox.ShutdownAfter)
	}

	// Create analyzer with whitelist filter and proxmox client
	anlz := analyzer.New(analyzer.Config{
		AgentURL:     cfg.Analyzer.AgentURL,
		Timeout:      cfg.Analyzer.Timeout,
		AgentEnabled: cfg.Analyzer.Enabled,
	}, whitelistFilter, proxmoxClient)

	if cfg.Analyzer.Enabled {
		log.Println("Agent enabled - will send samples to remote agent")

		// Check agent health
		if err := anlz.CheckAgentHealth(); err != nil {
			log.Printf("WARNING: Agent health check failed: %v", err)
			log.Printf("Uploads will fail until agent is reachable at: %s", cfg.Analyzer.AgentURL)
		} else {
			log.Println("Agent health check passed")
		}
	} else {
		log.Println("Agent disabled - will create stub reports (set analyzer.enabled=true in config)")
	}

	// Create handler
	handler := api.NewHandler(repo, anlz, cfg.Data.SamplesDir)

	// Setup routes
	router := api.SetupRoutes(handler, cfg.Server.FrontendDir)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("CupaX server starting on http://localhost%s", addr)
	log.Println("Press Ctrl+C to stop")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
