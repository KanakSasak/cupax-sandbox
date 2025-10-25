package api

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cupax/cupax/internal/analyzer"
	"github.com/cupax/cupax/internal/database"
	"github.com/cupax/cupax/internal/models"
	"github.com/cupax/cupax/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler contains all API handlers
type Handler struct {
	repo       *database.Repository
	analyzer   *analyzer.Analyzer
	samplesDir string
}

// NewHandler creates a new handler instance
func NewHandler(repo *database.Repository, anlz *analyzer.Analyzer, samplesDir string) *Handler {
	return &Handler{
		repo:       repo,
		analyzer:   anlz,
		samplesDir: samplesDir,
	}
}

// UploadFileResponse represents the file upload response
type UploadFileResponse struct {
	AnalysisID uuid.UUID `json:"analysis_id"`
	Message    string    `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleUploadFile handles file upload and creates an analysis job
func (h *Handler) HandleUploadFile(c *gin.Context) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No file uploaded"})
		return
	}
	defer file.Close()

	// Get optional parameters
	isZip := c.PostForm("is_zip") == "true"
	zipPassword := c.PostForm("zip_password")

	// Auto-detect zip files by extension
	if !isZip && filepath.Ext(header.Filename) == ".zip" {
		isZip = true
		// Use default password if not provided
		if zipPassword == "" {
			zipPassword = "infected"
		}
		log.Printf("Auto-detected zip file: %s, using password: %s", header.Filename, zipPassword)
	}

	// Validate file size (max 100MB)
	if header.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "File size exceeds 100MB limit"})
		return
	}

	// Calculate SHA256 hash
	hash, err := utils.CalculateSHA256FromReader(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate file hash"})
		return
	}

	// Reset file pointer for saving
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process file"})
		return
	}

	// Check if file already exists by hash
	existing, err := h.repo.GetAnalysisByHash(hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	if existing != nil {
		c.JSON(http.StatusOK, UploadFileResponse{
			AnalysisID: existing.ID,
			Message:    "File already analyzed. Returning existing analysis.",
		})
		return
	}

	// Ensure samples directory exists
	if err := os.MkdirAll(h.samplesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create samples directory"})
		return
	}

	// Store sample with hash as filename, preserve original extension
	fileExt := filepath.Ext(header.Filename)
	sampleFilename := hash + fileExt
	samplePath := filepath.Join(h.samplesDir, sampleFilename)

	log.Printf("Saving sample: %s (original: %s, ext: %s)", sampleFilename, header.Filename, fileExt)

	// Save file to disk
	dst, err := os.Create(samplePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file"})
		return
	}

	// Create analysis record
	analysisID := uuid.New()
	analysis := &models.Analysis{
		ID:             analysisID,
		Filename:       header.Filename,
		FileHashSHA256: hash,
		Status:         models.StatusRunning,
		SubmittedAt:    time.Now(),
	}

	if err := h.repo.CreateAnalysis(analysis); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create analysis record"})
		return
	}

	// Execute analysis synchronously via remote agent
	log.Printf("Starting analysis for file: %s (ID: %s, is_zip: %v)", header.Filename, analysisID, isZip)
	result := h.analyzer.Analyze(samplePath, analysisID.String(), isZip, zipPassword)

	if result.Error != nil {
		// Update analysis status to error
		log.Printf("Analysis failed: %v", result.Error)
		errMsg := result.Error.Error()
		analysis.Status = models.StatusError
		analysis.ErrorMessage = &errMsg
		analysis.CompletedAt = &[]time.Time{time.Now()}[0]
		h.repo.UpdateAnalysisStatus(analysisID, analysis.Status, analysis.ErrorMessage)

		c.JSON(http.StatusOK, UploadFileResponse{
			AnalysisID: analysisID,
			Message:    "File uploaded but analysis failed. Check analysis details for errors.",
		})
		return
	}

	// Update analysis with results
	log.Printf("Analysis completed successfully for: %s", header.Filename)
	if err := h.repo.UpdateAnalysisReport(analysisID, result.Report); err != nil {
		log.Printf("Failed to save analysis report: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Analysis completed but failed to save results"})
		return
	}

	c.JSON(http.StatusCreated, UploadFileResponse{
		AnalysisID: analysisID,
		Message:    "File uploaded and analyzed successfully.",
	})
}

// HandleGetAnalyses returns all analyses
func (h *Handler) HandleGetAnalyses(c *gin.Context) {
	analyses, err := h.repo.GetAllAnalyses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve analyses"})
		return
	}

	// Return empty array instead of null if no analyses
	if analyses == nil {
		analyses = []models.Analysis{}
	}

	c.JSON(http.StatusOK, analyses)
}

// HandleGetAnalysisByID returns a specific analysis
func (h *Handler) HandleGetAnalysisByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid analysis ID"})
		return
	}

	analysis, err := h.repo.GetAnalysisByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Analysis not found"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
