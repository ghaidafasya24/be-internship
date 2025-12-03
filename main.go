package main

import (
	"be-internship/config"
	route "be-internship/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Tidak dapat memuat .env, menggunakan environment variable sistem...")
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New(config.Cors))

	route.SetupRoutes(app)

	log.Printf("🚀 Server is running on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
