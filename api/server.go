package main

import (
	"fmt"
	"htmlfmt/internal/ai"
	"htmlfmt/internal/analyzer"
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
