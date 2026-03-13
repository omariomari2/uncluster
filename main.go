package main

import (
	"fmt"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/converter"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/formatter"
	"htmlfmt/internal/nodejs"
	"htmlfmt/internal/scraper"
	"htmlfmt/internal/zipper"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50 MB — allows large ZIP uploads and scraped pages
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	setupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if err := app.Listen(":" + port); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

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

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/format", handleFormat)

	api.Post("/convert", handleConvert)

	api.Post("/analyze", handleAnalyze)

	api.Post("/export", handleExport)

	api.Post("/export-nodejs", handleExportNodeJS)

	api.Post("/export-nodejs-ejs", handleExportNodeJSEJS)

	api.Post("/scrape", handleScrape)
	api.Post("/scrape-nodejs", handleScrapeNodeJS)
	api.Post("/scrape-nodejs-ejs", handleScrapeNodeJSEJS)

	api.Get("/health", handleHealth)

	app.Static("/", "./dist")
}

func handleFormat(c *fiber.Ctx) error {
	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

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

func handleConvert(c *fiber.Ctx) error {
	var req ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	jsx, err := converter.ConvertToJSX(req.HTML, "", "", nil, nil)
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

func handleAnalyze(c *fiber.Ctx) error {
	var req ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ComponentResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(ComponentResponse{
			Success: false,
			Error:   "HTML content is required",
		})
	}

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

func handleExport(c *fiber.Ctx) error {
	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	extracted, err := extractor.Extract(req.HTML)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	zipData, err := zipper.CreateZipWithMetadata(extracted.HTML, extracted.InlineCSS, extracted.InlineJS, extracted.ExternalCSS, extracted.ExternalJS, extracted.LocalAssets)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\"extracted.zip\"")
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))

	return c.Send(zipData)
}

func handleExportNodeJS(c *fiber.Ctx) error {
	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	extracted, err := extractor.Extract(req.HTML)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	rewrittenHTML := extracted.RewriteForNodeJS()

	projectName := fmt.Sprintf("project-%d", time.Now().Unix())

	config := &nodejs.ProjectConfig{
		ProjectName:    projectName,
		PackageManager: "npm",
		HTML:           rewrittenHTML,
		CSS:            extracted.CSS,
		JS:             extracted.JS,
		ExternalCSS:    extracted.ExternalCSS,
		ExternalJS:     extracted.ExternalJS,
	}

	projectFiles, err := nodejs.GenerateProject(config)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	zipData, err := nodejs.CreateProjectZip(projectFiles.Files, projectName)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", projectName))
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))

	return c.Send(zipData)
}

func handleExportNodeJSEJS(c *fiber.Ctx) error {
	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.Status(400).JSON(Response{
			Success: false,
			Error:   "HTML content is required",
		})
	}

	extracted, err := extractor.Extract(req.HTML)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	projectName := fmt.Sprintf("project-%d", time.Now().Unix())

	rewrittenHTML := extracted.RewriteForEJS()

	config := &nodejs.EJSProjectConfig{
		ProjectName: projectName,
		HTML:        rewrittenHTML,
		InlineCSS:   extracted.InlineCSS,
		InlineJS:    extracted.InlineJS,
		ExternalCSS: extracted.ExternalCSS,
		ExternalJS:  extracted.ExternalJS,
	}

	projectFiles, err := nodejs.GenerateEJSProject(config)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	zipData, err := nodejs.CreateProjectZip(projectFiles.Files, projectName)
	if err != nil {
		return c.Status(500).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-ejs.zip\"", projectName))
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))

	return c.Send(zipData)
}

type ScrapeRequest struct {
	URL string `json:"url"`
}

func handleScrape(c *fiber.Ctx) error {
	var req ScrapeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{Success: false, Error: "Invalid request body"})
	}
	if strings.TrimSpace(req.URL) == "" {
		return c.Status(400).JSON(Response{Success: false, Error: "URL is required"})
	}

	extracted, err := scraper.ScrapeURL(req.URL)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	zipData, err := zipper.CreateZipWithMetadata(extracted.HTML, extracted.InlineCSS, extracted.InlineJS, extracted.ExternalCSS, extracted.ExternalJS, extracted.LocalAssets)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\"extracted.zip\"")
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
	return c.Send(zipData)
}

func handleScrapeNodeJS(c *fiber.Ctx) error {
	var req ScrapeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{Success: false, Error: "Invalid request body"})
	}
	if strings.TrimSpace(req.URL) == "" {
		return c.Status(400).JSON(Response{Success: false, Error: "URL is required"})
	}

	extracted, err := scraper.ScrapeURL(req.URL)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	rewrittenHTML := extracted.RewriteForNodeJS()
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())

	config := &nodejs.ProjectConfig{
		ProjectName:    projectName,
		PackageManager: "npm",
		HTML:           rewrittenHTML,
		CSS:            extracted.CSS,
		JS:             extracted.JS,
		ExternalCSS:    extracted.ExternalCSS,
		ExternalJS:     extracted.ExternalJS,
	}

	projectFiles, err := nodejs.GenerateProject(config)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	binaryFiles := make(map[string][]byte, len(extracted.LocalAssets))
	for _, asset := range extracted.LocalAssets {
		binaryFiles["public/"+asset.Path] = asset.Content
	}

	zipData, err := nodejs.CreateProjectZipWithBinary(projectFiles.Files, binaryFiles, projectName)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", projectName))
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
	return c.Send(zipData)
}

func handleScrapeNodeJSEJS(c *fiber.Ctx) error {
	var req ScrapeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(Response{Success: false, Error: "Invalid request body"})
	}
	if strings.TrimSpace(req.URL) == "" {
		return c.Status(400).JSON(Response{Success: false, Error: "URL is required"})
	}

	extracted, err := scraper.ScrapeURL(req.URL)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	rewrittenHTML := extracted.RewriteForEJS()
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())

	config := &nodejs.EJSProjectConfig{
		ProjectName: projectName,
		HTML:        rewrittenHTML,
		InlineCSS:   extracted.InlineCSS,
		InlineJS:    extracted.InlineJS,
		ExternalCSS: extracted.ExternalCSS,
		ExternalJS:  extracted.ExternalJS,
	}

	projectFiles, err := nodejs.GenerateEJSProject(config)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	binaryFiles := make(map[string][]byte, len(extracted.LocalAssets))
	for _, asset := range extracted.LocalAssets {
		binaryFiles["public/"+asset.Path] = asset.Content
	}

	zipData, err := nodejs.CreateProjectZipWithBinary(projectFiles.Files, binaryFiles, projectName)
	if err != nil {
		return c.Status(500).JSON(Response{Success: false, Error: err.Error()})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-ejs.zip\"", projectName))
	c.Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
	return c.Send(zipData)
}

func handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "htmlfmt-api",
		"version": "1.0.0",
	})
}
