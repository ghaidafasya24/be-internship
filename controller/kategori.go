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

// INSERT KATEGORI (pakai form-data + wajib token)
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

// // ✅ Fungsi untuk menambahkan kategori baru
// func InsertCategory(db *mongo.Database, col string, kategori model.Kategori) (primitive.ObjectID, error) {
// 	// Membuat dokumen BSON untuk disimpan ke MongoDB
// 	categoryData := bson.M{
// 		"nama_kategori": kategori.NamaKategori,
// 		"deskripsi":     kategori.Deskripsi,
// 	}

// 	// Menyisipkan dokumen ke koleksi
// 	result, err := db.Collection(col).InsertOne(context.Background(), categoryData)
// 	if err != nil {
// 		fmt.Printf("InsertCategory error: %v\n", err)
// 		return primitive.NilObjectID, err
// 	}

// 	insertedID := result.InsertedID.(primitive.ObjectID)
// 	return insertedID, nil
// }

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
		"total":   len(categories),
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

// UPDATE KATEGORI (Wajib Token, pakai form-data)
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

// DELETE KATEGORI (Wajib Token)
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
