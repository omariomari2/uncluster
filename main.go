package main

import (
	"fmt"
	"htmlfmt/internal/ai"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/converter"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/formatter"
	"htmlfmt/internal/nodejs"
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
	initCloudflareAI()

	app := fiber.New(fiber.Config{
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

func initCloudflareAI() {
	workerURL := os.Getenv("CLOUDFLARE_WORKER_URL")
	workerToken := os.Getenv("CLOUDFLARE_WORKER_TOKEN")
	workerModel := os.Getenv("CLOUDFLARE_WORKER_MODEL")
	if workerURL != "" {
		config := ai.WorkerAIConfig{
			URL:     workerURL,
			Token:   workerToken,
			Model:   workerModel,
			Enabled: true,
		}
		client := ai.NewWorkerAIClient(config)
		analyzer.SetAIClient(client)
		return
	}

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	model := os.Getenv("CLOUDFLARE_AI_MODEL")

	if accountID == "" || apiToken == "" {
		return
	}

	if model == "" {
		model = "@cf/meta/llama-3-8b-instruct"
	}

	config := ai.CloudflareConfig{
		AccountID: accountID,
		APIToken:  apiToken,
		Model:     model,
		Enabled:   true,
	}

	client := ai.NewCloudflareClient(config)
	analyzer.SetAIClient(client)

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

	zipData, err := zipper.CreateZipWithMetadata(extracted.HTML, extracted.InlineCSS, extracted.InlineJS, extracted.ExternalCSS, extracted.ExternalJS)
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

func handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "htmlfmt-api",
		"version": "1.0.0",
	})
}
