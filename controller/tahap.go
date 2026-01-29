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

// InsertTahap godoc
// @Summary      Insert Tahap
// @Description  Menambahkan data tahap penyimpanan baru ke dalam sistem.
// @Tags         Data Tempat Penyimpanan (Tahap)
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        nama_tahap  formData  string  true  "Nama Tahap"
// @Success      201  {object}  map[string]interface{}  "Data tahap berhasil ditambahkan"
// @Router       /tahap [post]
func InsertTahap(c *fiber.Ctx) error {
	// 🔹 Ambil value dari form-data
	namaTahap := c.FormValue("nama_tahap")
	// 🔹 Validasi field wajib
	if namaTahap == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama tahap tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tahapCollection := config.Ulbimongoconn.Collection("tahap")

	// 🔹 Cek apakah kategori sudah ada berdasarkan nama
	var existing model.Tahap
	err := tahapCollection.FindOne(ctx, bson.M{
		"nama_tahap": namaTahap,
	}).Decode(&existing)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Data Tahap sudah terdaftar",
		})
	}

	// 🔹 Buat data kategori baru
	newTahap := model.Tahap{
		ID:        primitive.NewObjectID(),
		NamaTahap: namaTahap,
	}

	// 🔹 Insert ke database
	_, err = tahapCollection.InsertOne(ctx, newTahap)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan data tahap ke database",
		})
	}

	// 🔹 Response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Data Tahap berhasil ditambahkan",
		"data":    newTahap,
	})
}

// UpdateTahapByID godoc
// @Summary      Update Tahap
// @Description  Memperbarui data tahap penyimpanan berdasarkan ID tahap
// @Tags         Data Tempat Penyimpanan (Tahap)
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      string  true  "ID Tahap"
// @Param        nama_tahap  formData  string  true  "Nama Tahap"
// @Success      200 {object} map[string]interface{} "Data tahap berhasil diperbarui"
// @Router       /tahap/{id} [put]
func UpdateTahapByID(c *fiber.Ctx) error {
	// =========================
	// VALIDASI ID
	// =========================
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tahap tidak valid",
		})
	}

	// =========================
	// AMBIL FORM DATA
	// =========================
	namaTahap := c.FormValue("nama_tahap")
	if namaTahap == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama tahap tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tahapCollection := config.Ulbimongoconn.Collection("tahap")

	// =========================
	// CEK APAKAH DATA ADA
	// =========================
	var existing model.Tahap
	err = tahapCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data tahap tidak ditemukan",
		})
	}

	// =========================
	// PROSES UPDATE
	// =========================
	update := bson.M{
		"$set": bson.M{
			"nama_tahap": namaTahap,
		},
	}

	_, err = tahapCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memperbarui data tahap",
		})
	}

	// =========================
	// RESPONSE
	// =========================
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data tahap berhasil diperbarui",
		"data": fiber.Map{
			"_id":        objID,
			"nama_tahap": namaTahap,
		},
	})
}

// GetAllTahap godoc
// @Summary      Get All Tahap
// @Description  Mengambil seluruh data tahap penyimpanan dari database MongoDB.
// @Tags         Data Tempat Penyimpanan (Tahap)
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.GetAllTahapResponse  "Berhasil mengambil data tahap"
// @Router       /tahap [get]
func GetAllTahap(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	col := db.Collection("tahap") // nama koleksi MongoDB

	filter := bson.M{}
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error GetAllTahap:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal mengambil data tahap",
			"error":   err.Error(),
		})
	}

	var tahaps []model.Tahap
	err = cursor.All(context.TODO(), &tahaps)
	if err != nil {
		fmt.Println("Error decode:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal decode data tahap",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil semua data tahap",
		"data":    tahaps,
		"total":   len(tahaps),
	})
}

// GetTahapByID godoc
// @Summary      Get Tahap by ID
// @Description  Mengambil data tahap penyimpanan berdasarkan ID tahap
// @Tags         Data Tempat Penyimpanan (Tahap)
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID Tahap"
// @Success      200  {object}  model.Tahap  "Data tahap berhasil ditemukan"
// @Router       /tahap/{id} [get]
func GetTahapByID(c *fiber.Ctx) error {
	// Ambil parameter ID dari URL
	idParam := c.Params("id")

	// Konversi ID string menjadi ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tahap tidak valid",
		})
	}

	// Ambil koleksi tahap
	collection := config.Ulbimongoconn.Collection("tahap")

	// Query ke database
	var tahap model.Tahap
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&tahap)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data Tahap tidak ditemukan",
		})
	}

	// Return hasil
	return c.Status(fiber.StatusOK).JSON(tahap)
}

// DeleteTahapByID godoc
// @Summary      Delete Tahap by ID
// @Description  Menghapus data tahap penyimpanan berdasarkan ID tahap
// @Tags         Data Tempat Penyimpanan (Tahap)
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID Tahap"
// @Success      200  {object}  map[string]string "Data tahap berhasil dihapus"
// @Router       /tahap/{id} [delete]
func DeleteTahapByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tahap tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tahapCollection := config.Ulbimongoconn.Collection("tahap")

	// Hapus kategori
	result, err := tahapCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus data tahap",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data tahap tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data tahap berhasil dihapus",
		"id":      idParam,
	})
}
