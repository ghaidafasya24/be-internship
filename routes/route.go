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
	userRoutes.Post("/register", controller.Register)          // Route untuk registrasi pengguna
	userRoutes.Post("/login", controller.Login)                // Route untuk login pengguna
	userRoutes.Get("/", controller.GetAllUsers)                // Route untuk mengambil data pengguna
	userRoutes.Get("/:username", controller.GetUserByUsername) // Route untuk mengambil data pengguna berdasarkan username
	// userRoutes.Get("/:phone_number", controller.GetUserByPhoneNumber)        // Route untuk mengambil data pengguna berdasarkan nomor telepon
	userRoutes.Get("/:id", controller.GetUserByID)                           // Route untuk mengambil data pengguna berdasarkan ID
	userRoutes.Put("/:id", controller.JWTAuth, controller.UpdateUserByID)    // Route untuk mengupdate data pengguna berdasarkan ID
	userRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteUserByID) // Route untuk menghapus data pengguna berdasarkan ID
	// app.Post("/auth/request-reset", controller.RequestResetPassword)
	// app.Post("/auth/reset-password")

	// Koleksi routes
	koleksiRoutes := api.Group("/koleksi")
	koleksiRoutes.Post("/", controller.JWTAuth, controller.InsertKoleksi)
	koleksiRoutes.Get("/", controller.GetAllKoleksi)
	koleksiRoutes.Get("/:id", controller.GetKoleksiByID)
	koleksiRoutes.Put("/:id", controller.JWTAuth, controller.UpdateKoleksi)
	koleksiRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteKoleksiByID)

	// Kategori routes
	kategoriRoutes := api.Group("/kategori")
	kategoriRoutes.Post("/", controller.JWTAuth, controller.InsertKategori)
	kategoriRoutes.Get("/", controller.GetAllCategory)
	kategoriRoutes.Get("/:id", controller.GetCategoryByID)
	kategoriRoutes.Put("/:id", controller.JWTAuth, controller.UpdateKategori)
	kategoriRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteKategoriByID)

	// Gudang routes
	GudangRoutes := api.Group("/gudang")
	GudangRoutes.Post("/", controller.JWTAuth, controller.InsertGudang)
	GudangRoutes.Get("/", controller.GetAllGudang)
	GudangRoutes.Get("/:id", controller.GetGudangByID)
	GudangRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteGudangByID)

	// Rak routes
	RakRoutes := api.Group("/rak")
	RakRoutes.Post("/", controller.JWTAuth, controller.InsertRak)
	RakRoutes.Get("/", controller.GetAllRak)
	RakRoutes.Get("/:id", controller.GetRakByID)
	RakRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteRakByID)

	// Tahap routes
	TahapRoutes := api.Group("/tahap")
	TahapRoutes.Post("/", controller.JWTAuth, controller.InsertTahap)
	TahapRoutes.Get("/", controller.GetAllTahap)
	TahapRoutes.Get("/:id", controller.GetTahapByID)
	TahapRoutes.Delete("/:id", controller.JWTAuth, controller.DeleteTahapByID)
}
