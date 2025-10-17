package main

import (
	"fmt"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/converter"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/formatter"
	"htmlfmt/internal/zipper"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Request structures
type FormatRequest struct {
	HTML string `json:"html" validate:"required"`
}

type ConvertRequest struct {
	HTML string `json:"html" validate:"required"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ComponentResponse struct {
	Success     bool                           `json:"success"`
	Suggestions []analyzer.ComponentSuggestion `json:"suggestions,omitempty"`
	Error       string                         `json:"error,omitempty"`
}

// setupRoutes configures all API routes
func setupRoutes(app *fiber.App) {
	// API routes
	api := app.Group("/api")

	// Format HTML endpoint
	api.Post("/format", handleFormat)

	// Convert to JSX endpoint
	api.Post("/convert", handleConvert)

	// Analyze components endpoint
	api.Post("/analyze", handleAnalyze)

	// Export to zip endpoint
	api.Post("/export", handleExport)

	// Health check
	api.Get("/health", handleHealth)

	// Serve static files
	app.Static("/", "./web/static")
}

// handleFormat formats HTML with proper indentation
func handleFormat(c *fiber.Ctx) error {
	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate input
	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	// Format HTML
	formatted, err := formatter.Format(req.HTML)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Success: true,
		Data:    formatted,
	})
}

// handleConvert converts HTML to JSX
func handleConvert(c *fiber.Ctx) error {
	var req ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate input
	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	// Convert to JSX
	jsx, err := converter.ConvertToJSX(req.HTML)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Success: true,
		Data:    jsx,
	})
}

// handleAnalyze analyzes HTML and returns component suggestions
func handleAnalyze(c *fiber.Ctx) error {
	var req ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ComponentResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate input
	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(ComponentResponse{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	// Analyze components
	suggestions, err := analyzer.AnalyzeComponents(req.HTML)
	if err != nil {
		return c.Status(500).JSON(ComponentResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(ComponentResponse{
		Success:     true,
		Suggestions: suggestions,
	})
}

// handleExport extracts CSS and JS from HTML and returns a zip file
func handleExport(c *fiber.Ctx) error {
	log.Printf("📦 Export request received from %s", c.IP())

	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Export request parsing failed: %v", err)
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate input
	if strings.TrimSpace(req.HTML) == "" {
		log.Printf("❌ Export request: empty HTML content")
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	log.Printf("📄 Extracting CSS/JS from HTML (length: %d chars)", len(req.HTML))
	// Extract CSS and JS from HTML
	extracted, err := extractor.Extract(req.HTML)
	if err != nil {
		log.Printf("❌ Extraction failed: %v", err)
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	log.Printf("📊 Extraction results - HTML: %d chars, CSS: %d chars, JS: %d chars",
		len(extracted.HTML), len(extracted.CSS), len(extracted.JS))
	log.Printf("📦 External resources - CSS: %d files, JS: %d files",
		len(extracted.ExternalCSS), len(extracted.ExternalJS))

	// Create zip archive
	log.Printf("🗜️ Creating zip archive...")
	zipData, err := zipper.CreateZipWithMetadata(extracted.HTML, extracted.CSS, extracted.JS, extracted.ExternalCSS, extracted.ExternalJS)
	if err != nil {
		log.Printf("❌ Zip creation failed: %v", err)
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Set headers for file download
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\"extracted.zip\"")
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))

	log.Printf("✅ Export completed successfully (zip size: %d bytes)", len(zipData))
	// Return the zip file
	return c.Send(zipData)
}

// handleHealth returns server health status
func handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "htmlfmt-api",
		"version": "1.0.0",
	})
}
