package main

import (
	"htmlfmt/internal/ai"
	"htmlfmt/internal/analyzer"
	"log"
	"os"

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

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func initCloudflareAI() {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	model := os.Getenv("CLOUDFLARE_AI_MODEL")

	if accountID == "" || apiToken == "" {
		log.Printf("ℹ️  Cloudflare AI not configured (CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN required)")
		log.Printf("ℹ️  Component analysis will use pattern-based detection only")
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

	log.Printf("✅ Cloudflare AI initialized (Model: %s)", model)
	log.Printf("🤖 AI-powered component analysis is enabled")
}
