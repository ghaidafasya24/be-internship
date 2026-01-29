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

// InsertGudang godoc
// @Summary      Insert Gudang
// @Description  Menambahkan data gudang baru ke dalam sistem
// @Tags         Data Tempat Penyimpanan (Gudang)
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        nama_gudang  formData  string  true  "Nama Gudang"
// @Success      201 {object} map[string]interface{} "Data gudang berhasil ditambahkan"
// @Router       /gudang [post]
func InsertGudang(c *fiber.Ctx) error {
	// 🔹 Ambil value dari form-data
	namaGudang := c.FormValue("nama_gudang")
	// 🔹 Validasi field wajib
	if namaGudang == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama gudang tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gudangCollection := config.Ulbimongoconn.Collection("gudang")

	// 🔹 Cek apakah kategori sudah ada berdasarkan nama
	var existing model.Gudang
	err := gudangCollection.FindOne(ctx, bson.M{
		"nama_gudang": namaGudang,
	}).Decode(&existing)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Data Gudang sudah terdaftar",
		})
	}

	// 🔹 Buat data kategori baru
	newGudang := model.Gudang{
		ID:         primitive.NewObjectID(),
		NamaGudang: namaGudang,
	}

	// 🔹 Insert ke database
	_, err = gudangCollection.InsertOne(ctx, newGudang)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan data gudang ke database",
		})
	}

	// 🔹 Response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Data Gudang berhasil ditambahkan",
		"data":    newGudang,
	})
}

// UpdateGudangByID godoc
// @Summary      Update Gudang by ID
// @Description  Memperbarui data gudang berdasarkan ID
// @Tags         Data Tempat Penyimpanan (Gudang)
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string  true  "ID Gudang"
// @Param        nama_gudang  formData  string  true  "Nama Gudang"
// @Success      200 {object} map[string]interface{} "Data gudang berhasil diperbarui"
// @Router       /gudang/{id} [put]
func UpdateGudangByID(c *fiber.Ctx) error {
	// =========================
	// VALIDASI ID
	// =========================
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID gudang tidak valid",
		})
	}

	// =========================
	// AMBIL FORM DATA
	// =========================
	namaGudang := c.FormValue("nama_gudang")
	if namaGudang == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama gudang tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gudangCollection := config.Ulbimongoconn.Collection("gudang")

	// =========================
	// CEK APAKAH DATA ADA
	// =========================
	var existing model.Gudang
	err = gudangCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data gudang tidak ditemukan",
		})
	}

	// =========================
	// PROSES UPDATE
	// =========================
	update := bson.M{
		"$set": bson.M{
			"nama_gudang": namaGudang,
		},
	}

	_, err = gudangCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memperbarui data gudang",
		})
	}

	// =========================
	// RESPONSE
	// =========================
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data gudang berhasil diperbarui",
		"data": fiber.Map{
			"_id":         objID,
			"nama_gudang": namaGudang,
		},
	})
}

// GetAllGudang godoc
// @Summary      Get All Gudang
// @Description  Mengambil seluruh data gudang dari database MongoDB
// @Tags         Data Tempat Penyimpanan (Gudang)
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Berhasil mengambil semua data gudang"
// @Router       /gudang [get]
func GetAllGudang(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	col := db.Collection("gudang") // nama koleksi MongoDB

	filter := bson.M{}
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error GetAllGudang:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal mengambil data gudang",
			"error":   err.Error(),
		})
	}

	var gudangs []model.Gudang
	err = cursor.All(context.TODO(), &gudangs)
	if err != nil {
		fmt.Println("Error decode:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal decode data gudang",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil semua data gudang",
		"data":    gudangs,
	})
}

// GetGudangByID godoc
// @Summary      Get Gudang by ID
// @Description  Mengambil data gudang berdasarkan ID
// @Tags         Data Tempat Penyimpanan (Gudang)
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID Gudang"
// @Success      200  {object}  model.Gudang  "Data gudang berhasil ditemukan"
// @Router       /gudang/{id} [get]
func GetGudangByID(c *fiber.Ctx) error {
	// Ambil parameter ID dari URL
	idParam := c.Params("id")

	// Konversi ID string menjadi ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID gudang tidak valid",
		})
	}

	// Ambil koleksi gudang
	collection := config.Ulbimongoconn.Collection("gudang")

	// Query ke database
	var gudang model.Gudang
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&gudang)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data Gudang tidak ditemukan",
		})
	}

	// Return hasil
	return c.Status(fiber.StatusOK).JSON(gudang)
}

// DeleteGudangByID godoc
// @Summary      Delete Gudang by ID
// @Description  Menghapus data gudang berdasarkan ID
// @Tags         Data Tempat Penyimpanan (Gudang)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID Gudang"
// @Success      200  {object}  map[string]string "Data gudang berhasil dihapus"
// @Router       /gudang/{id} [delete]
func DeleteGudangByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID gudang tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gudangCollection := config.Ulbimongoconn.Collection("gudang")

	// Hapus kategori
	result, err := gudangCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus data gudang",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data gudang tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data gudang berhasil dihapus",
		"id":      idParam,
	})
}
