package main

import (
	"be-internship/config"
	_ "be-internship/docs"
	route "be-internship/routes"
	"log"
	"os"

	// alias untuk Swagger handler
	// alias untuk swagger files

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

// @title           API Koleksi Museum
// @version         1.0
// @description     Dokumentasi API untuk sistem pengelolaan museum

// @contact.name API Support
// @contact.url https://github.com/ghaidafasya24
// @contact.email 714220031@std.ulbi.ac.id

// @host            inventorymuseum-de54c3e9b901.herokuapp.com
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: Bearer <JWT Token>
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
