package controller

import (
	"be-internship/config"
	"be-internship/model"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertKategori godoc
// @Summary      Insert Kategori
// @Description  Menambahkan data kategori museum menggunakan form-data (wajib token)
// @Tags         Kategori
// @Accept       multipart/form-data
// @Produce      json
// @Param        nama_kategori  formData  string  true   "Nama kategori"
// @Param        deskripsi      formData  string  false  "Deskripsi kategori"
// @Success      201  {object}  map[string]interface{}
// @Router       /kategori [post]
// @Security     BearerAuth
func InsertKategori(c *fiber.Ctx) error {

	// 🔹 Ambil value dari form-data
	namaKategori := c.FormValue("nama_kategori")
	deskripsi := c.FormValue("deskripsi")

	// 🔹 Validasi field wajib
	if namaKategori == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama kategori tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kategoriCollection := config.Ulbimongoconn.Collection("kategori")

	// 🔹 Cek apakah kategori sudah ada berdasarkan nama
	var existing model.Kategori
	err := kategoriCollection.FindOne(ctx, bson.M{
		"nama_kategori": namaKategori,
	}).Decode(&existing)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Kategori sudah terdaftar",
		})
	}

	// 🔹 Buat data kategori baru
	newKategori := model.Kategori{
		ID:           primitive.NewObjectID(),
		NamaKategori: namaKategori,
		Deskripsi:    deskripsi,
	}

	// 🔹 Insert ke database
	_, err = kategoriCollection.InsertOne(ctx, newKategori)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan kategori ke database",
		})
	}

	// 🔹 Response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Kategori berhasil ditambahkan",
		"data":    newKategori,
	})
}

// GetAllKategori godoc
// @Summary      Get All Kategori
// @Description  Mengambil semua data kategori koleksi
// @Tags         Kategori
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /kategori [get]
func GetAllCategory(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	col := db.Collection("kategori") // nama koleksi MongoDB

	filter := bson.M{}
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error GetAllCategory:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal mengambil data kategori",
			"error":   err.Error(),
		})
	}

	var categories []model.Kategori
	err = cursor.All(context.TODO(), &categories)
	if err != nil {
		fmt.Println("Error decode:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal decode data kategori",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil semua data kategori",
		"total":   len(categories),
		"data":    categories,
	})
}

// GetKategoriByID godoc
// @Summary      Get Kategori by ID
// @Description  Mengambil satu data kategori koleksi berdasarkan ID
// @Tags         Kategori
// @Produce      json
// @Param        id   path      string  true  "ID Kategori"
// @Success      200  {object}  map[string]interface{}
// @Router       /kategori/{id} [get]
func GetCategoryByID(c *fiber.Ctx) error {
	// Ambil parameter ID dari URL
	idParam := c.Params("id")

	// Konversi ID string menjadi ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID kategori tidak valid",
		})
	}

	// Ambil koleksi kategori
	collection := config.Ulbimongoconn.Collection("kategori")

	// Query ke database
	var kategori model.Kategori
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&kategori)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Kategori tidak ditemukan",
		})
	}

	// Return hasil dengan pesan sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori dengan ID " + idParam + " berhasil ditampilkan",
		"data":    kategori,
	})
}

// UpdateKategori godoc
// @Summary      Update Kategori
// @Description  Mengubah data kategori berdasarkan ID. Endpoint ini memerlukan autentikasi JWT Bearer dan menggunakan form-data.
// @Tags         Kategori
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id             path      string  true   "ID Kategori"
// @Param        nama_kategori  formData  string  true   "Nama kategori"
// @Param        deskripsi      formData  string  false  "Deskripsi kategori"
// @Success      200  {object}  map[string]interface{} "Kategori berhasil diperbarui"
// @Router       /kategori/{id} [put]
func UpdateKategori(c *fiber.Ctx) error {

	// ============================
	// 1. Ambil ID dari URL
	// ============================
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID kategori wajib diisi",
		})
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID kategori tidak valid",
		})
	}

	// ============================
	// 2. Ambil data dari form-data
	// ============================
	namaKategori := c.FormValue("nama_kategori")
	deskripsi := c.FormValue("deskripsi")

	// Validasi
	if namaKategori == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama kategori tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kategoriCollection := config.Ulbimongoconn.Collection("kategori")

	// ============================
	// 3. Periksa apakah kategori ada
	// ============================
	var existing model.Kategori
	err = kategoriCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Kategori tidak ditemukan",
		})
	}

	// ============================
	// 4. Cek apakah nama sudah dipakai kategori lain
	// ============================
	var checkName model.Kategori
	err = kategoriCollection.FindOne(ctx, bson.M{
		"nama_kategori": namaKategori,
		"_id":           bson.M{"$ne": objID},
	}).Decode(&checkName)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Nama kategori sudah digunakan kategori lain",
		})
	}

	// ============================
	// 5. Update data
	// ============================
	update := bson.M{
		"$set": bson.M{
			"nama_kategori": namaKategori,
			"deskripsi":     deskripsi,
		},
	}

	_, err = kategoriCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengupdate data kategori",
		})
	}

	// Ambil ulang data setelah update
	var updated model.Kategori
	kategoriCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updated)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil diperbarui",
		"data":    updated,
	})
}

// DeleteKategoriByID godoc
// @Summary      Delete Kategori
// @Description  Menghapus data kategori berdasarkan ID (wajib autentikasi JWT Bearer)
// @Tags         Kategori
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID kategori"
// @Success      200  {object}  map[string]interface{}  "Kategori berhasil dihapus"
// @Router       /kategori/{id} [delete]
func DeleteKategoriByID(c *fiber.Ctx) error {

	// Ambil ID dari parameter
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID kategori tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kategoriCollection := config.Ulbimongoconn.Collection("kategori")

	// Hapus kategori
	result, err := kategoriCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus kategori",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Kategori tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil dihapus",
		"id":      idParam,
	})
}
