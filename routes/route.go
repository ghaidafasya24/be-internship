package routes

import (
	"be-internship/controller"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes initializes all the application routes
func SetupRoutes(app *fiber.App) {
	// Group API routes
	api := app.Group("/api")

	// User routes
	userRoutes := api.Group("/users")
	userRoutes.Post("/register", controller.Register) // Route untuk registrasi pengguna
	userRoutes.Post("/login", controller.Login)       // Route untuk login pengguna

	// Koleksi routes
	koleksiRoutes := api.Group("/koleksi")
	// Insert

	koleksiRoutes.Post("/", controller.InsertKoleksi)
	koleksiRoutes.Get("/", controller.GetAllKoleksi)
	koleksiRoutes.Get("/:id", controller.GetKoleksiByID)
	koleksiRoutes.Put("/:id", controller.UpdateKoleksi)
	koleksiRoutes.Delete("/:id", controller.DeleteKoleksiByID)

	// Tambahkan kategori route
	KategoriRoutes(api)
}
