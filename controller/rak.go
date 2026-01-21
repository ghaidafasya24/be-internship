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

// InsertRak godoc
// @Summary      Insert Rak
// @Description  Menambahkan data rak baru ke dalam sistem.
// @Tags         Data Tempat Penyimpanan
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        nama_rak  formData string true "Nama Rak"
// @Success      201 {object} map[string]interface{} "Data Rak berhasil ditambahkan"
// @Router       /rak [post]
func InsertRak(c *fiber.Ctx) error {
	// 🔹 Ambil value dari form-data
	namaRak := c.FormValue("nama_rak")
	// 🔹 Validasi field wajib
	if namaRak == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama rak tidak boleh kosong",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rakCollection := config.Ulbimongoconn.Collection("rak")

	// 🔹 Cek apakah kategori sudah ada berdasarkan nama
	var existing model.Rak
	err := rakCollection.FindOne(ctx, bson.M{
		"nama_rak": namaRak,
	}).Decode(&existing)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Data Rak sudah terdaftar",
		})
	}

	// 🔹 Buat data kategori baru
	newRak := model.Rak{
		ID:      primitive.NewObjectID(),
		NamaRak: namaRak,
	}

	// 🔹 Insert ke database
	_, err = rakCollection.InsertOne(ctx, newRak)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan data rak ke database",
		})
	}

	// 🔹 Response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Data Rak berhasil ditambahkan",
		"data":    newRak,
	})
}

// GetAllRak godoc
// @Summary      Get All Rak
// @Description  Mengambil seluruh data rak dari database MongoDB.
// @Tags         Data Tempat Penyimpanan
// @Accept       json
// @Produce      json
// @Success      200 {object} model.GetAllRakResponse "Success"
// @Router       /rak [get]
func GetAllRak(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	col := db.Collection("rak") // nama koleksi MongoDB

	filter := bson.M{}
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error GetAllRak:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal mengambil data rak",
			"error":   err.Error(),
		})
	}

	var raks []model.Rak
	err = cursor.All(context.TODO(), &raks)
	if err != nil {
		fmt.Println("Error decode:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal decode data rak",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil semua data rak",
		"data":    raks,
		"total":   len(raks),
	})
}

func GetRakByID(c *fiber.Ctx) error {
	// Ambil parameter ID dari URL
	idParam := c.Params("id")

	// Konversi ID string menjadi ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID rak tidak valid",
		})
	}

	// Ambil koleksi rak
	collection := config.Ulbimongoconn.Collection("rak")

	// Query ke database
	var rak model.Rak
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rak)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data Rak tidak ditemukan",
		})
	}

	// Return hasil
	return c.Status(fiber.StatusOK).JSON(rak)
}

func DeleteRakByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID rak tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rakCollection := config.Ulbimongoconn.Collection("rak")

	// Hapus kategori
	result, err := rakCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus data rak",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data rak tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data rak berhasil dihapus",
		"id":      idParam,
	})
}
