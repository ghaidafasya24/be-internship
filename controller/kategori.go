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
	"go.mongodb.org/mongo-driver/mongo"
)

// ✅ Fungsi untuk menambahkan kategori baru
func InsertCategory(db *mongo.Database, col string, kategori model.Kategori) (primitive.ObjectID, error) {
	// Membuat dokumen BSON untuk disimpan ke MongoDB
	categoryData := bson.M{
		"nama_kategori": kategori.NamaKategori,
		"deskripsi":     kategori.Deskripsi,
	}

	// Menyisipkan dokumen ke koleksi
	result, err := db.Collection(col).InsertOne(context.Background(), categoryData)
	if err != nil {
		fmt.Printf("InsertCategory error: %v\n", err)
		return primitive.NilObjectID, err
	}

	insertedID := result.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

// GetAllCategory mengambil semua data kategori dari MongoDB
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
		"data":    categories,
	})
}

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

	// Return hasil
	return c.Status(fiber.StatusOK).JSON(kategori)
}

// update kategori
func UpdateCategory(c *fiber.Ctx) error {
	// Ambil ID dari parameter URL
	idParam := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	// Ambil data dari body request
	var updateData model.Kategori
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Request tidak valid",
		})
	}

	// Buat map untuk field yang mau diupdate
	updateFields := bson.M{}

	if updateData.NamaKategori != "" {
		updateFields["nama_kategori"] = updateData.NamaKategori
	}

	if updateData.Deskripsi != "" {
		updateFields["deskripsi"] = updateData.Deskripsi
	}

	if len(updateFields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak ada field yang diupdate",
		})
	}

	// Proses update ke MongoDB
	db := config.Ulbimongoconn
	col := db.Collection("kategori")

	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateFields}

	result, err := col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println("UpdateCategory:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memperbarui kategori",
		})
	}

	if result.ModifiedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tidak ada data yang diubah dengan ID tersebut",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil diperbarui",
	})
}

func DeleteCategoryByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter URL
	idParam := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	// Koneksi ke database dan tentukan koleksi
	db := config.Ulbimongoconn
	col := db.Collection("kategori")

	// Filter berdasarkan ID
	filter := bson.M{"_id": id}

	// Hapus data
	result, err := col.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Gagal menghapus data untuk ID %s: %s", idParam, err.Error()),
		})
	}

	// Jika data tidak ditemukan
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Data dengan ID %s tidak ditemukan", idParam),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Kategori dengan ID %s berhasil dihapus", idParam),
	})
}
