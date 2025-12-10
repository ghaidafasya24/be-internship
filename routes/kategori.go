package routes

import (
	"be-internship/controller"
	"be-internship/model"

	"github.com/gofiber/fiber/v2"
)

func KategoriRoutes(router fiber.Router) {
	kategoriRoutes := router.Group("/kategori")

	//[POST] Tambah kategori
	kategoriRoutes.Post("/", func(c *fiber.Ctx) error {
		var data model.Kategori
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Request tidak valid",
			})
		}

		if data.NamaKategori == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Nama kategori wajib diisi",
			})
		}

		// id, err := controller.InsertCategory(config.Ulbimongoconn, "kategori", data)
		// if err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"error": "Gagal menambah kategori",
		// 	})
		// }

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Kategori berhasil ditambahkan",
			// "id":      id.Hex(),
		})
	})

	kategoriRoutes.Get("/", controller.GetAllCategory)
	kategoriRoutes.Get("/:id", controller.GetCategoryByID)
	// kategoriRoutes.Put("/:id", controller.UpdateCategory)
	// kategoriRoutes.Delete("/:id", controller.DeleteCategoryByID)
}
