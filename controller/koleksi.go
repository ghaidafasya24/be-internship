package controller

import (
	"be-internship/config"
	"be-internship/model"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fungsi utama untuk insert koleksi (pakai form-data)
func InsertKoleksi(c *fiber.Ctx) error {

	noReg := c.FormValue("no_reg")
	noInv := c.FormValue("no_inv")
	namaBenda := c.FormValue("nama_benda")
	tanggalPerolehan := c.FormValue("tanggal_perolehan")
	kategoriID := c.FormValue("kategori_id") // 🔹 ambil ID kategori, bukan nama
	bahan := c.FormValue("bahan")
	// ukuran := c.FormValue("ukuran")
	ukuran := model.Ukuran{
		ID:                 primitive.NewObjectID(),
		PanjangKeseluruhan: c.FormValue("panjang_keseluruhan"),
		Lebar:              c.FormValue("lebar"),
		Tebal:              c.FormValue("tebal"),
		Tinggi:             c.FormValue("tinggi"),
		Diameter:           c.FormValue("diameter"),
		Berat:              c.FormValue("berat"),
		// CreatedAt:          time.Now(), // ❗ WAJIB
	}
	asalKoleksi := c.FormValue("asal_koleksi")
	tempatPerolehan := c.FormValue("tempat_perolehan")
	deskripsi := c.FormValue("deskripsi")
	// tempatPenyimpanan := c.FormValue("tempat_penyimpanan")
	gudangID := c.FormValue("gudang_id") // 🔹 ambil ID gudang, bukan nama
	rakID := c.FormValue("rak_id")       // 🔹 ambil ID rak, bukan nama
	tahapID := c.FormValue("tahap_id")   // 🔹 ambil ID tahap, bukan nama
	Kondisi := c.FormValue("kondisi")

	if noReg == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No registrasi tidak boleh kosong.",
		})
	}

	if noInv == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No inventaris tidak boleh kosong.",
		})
	}

	if namaBenda == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama benda tidak boleh kosong.",
		})
	}

	// 🔹 Cek kategori berdasarkan ID
	objID, err := primitive.ObjectIDFromHex(kategoriID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID kategori tidak valid.",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var kategori model.Kategori
	kategoriCollection := config.Ulbimongoconn.Collection("kategori")
	err = kategoriCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&kategori)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Kategori tidak ditemukan.",
		})
	}

	// 🔹 Cek data gudang berdasarkan ID
	objID, err = primitive.ObjectIDFromHex(gudangID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID gudang tidak valid.",
		})
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var gudang model.Gudang
	gudangCollection := config.Ulbimongoconn.Collection("gudang")
	err = gudangCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&gudang)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data gudang tidak ditemukan.",
		})
	}

	// ======================================================
	// 🔹 RAK OPSIONAL
	// ======================================================
	var rak model.Rak
	if rakID != "" {
		objID, err = primitive.ObjectIDFromHex(rakID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ID rak tidak valid.",
			})
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		rakCollection := config.Ulbimongoconn.Collection("rak")
		err = rakCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&rak)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Data rak tidak ditemukan.",
			})
		}
	}

	// 🔹 Cek data tahap berdasarkan ID
	objID, err = primitive.ObjectIDFromHex(tahapID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tahap tidak valid.",
		})
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var tahap model.Tahap
	tahapCollection := config.Ulbimongoconn.Collection("tahap")
	err = tahapCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&tahap)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Data tahap tidak ditemukan.",
		})
	}

	// 🔹 Upload gambar OPSIONAL
	var imageURL string

	file, err := c.FormFile("foto")
	if err == nil && file != nil {
		// Jika ada file → upload ke GitHub
		imageURL, err = uploadImageToGitHub(file, namaBenda)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Gagal upload gambar ke GitHub: %v", err),
			})
		}
	} else {
		// Jika TIDAK ada file → kosongkan
		imageURL = ""
	}

	// 🔹 Upload gambar
	// file, err := c.FormFile("foto")
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "File foto wajib diunggah.",
	// 	})
	// }

	// imageURL, err := uploadImageToGitHub(file, namaBenda)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": fmt.Sprintf("Gagal upload gambar ke GitHub: %v", err),
	// 	})
	// }

	tempatPenyimpanan := model.TempatPenyimpanan{}

	if gudangID != "" {
		objGudangID, _ := primitive.ObjectIDFromHex(gudangID)
		tempatPenyimpanan.Gudang = model.Gudang{
			ID:         objGudangID,
			NamaGudang: gudang.NamaGudang,
		}
	}

	if rakID != "" {
		objRakID, _ := primitive.ObjectIDFromHex(rakID)
		tempatPenyimpanan.Rak = model.Rak{
			ID:      objRakID,
			NamaRak: rak.NamaRak,
		}
	}

	if tahapID != "" {
		objTahapID, _ := primitive.ObjectIDFromHex(tahapID)
		tempatPenyimpanan.Tahap = model.Tahap{
			ID:        objTahapID,
			NamaTahap: tahap.NamaTahap,
		}
	}

	// 🔹 Buat data koleksi
	data := model.Koleksi{
		ID:                primitive.NewObjectID(),
		Kategori:          kategori,
		NoRegistrasi:      noReg,
		NoInventaris:      noInv,
		NamaBenda:         namaBenda,
		AsalKoleksi:       asalKoleksi,
		Bahan:             bahan,
		Ukuran:            ukuran,
		TempatPerolehan:   tempatPerolehan,
		TanggalPerolehan:  tanggalPerolehan,
		Deskripsi:         deskripsi,
		TempatPenyimpanan: tempatPenyimpanan,
		Kondisi:           Kondisi,
		Foto:              imageURL,
		CreatedAt:         time.Now(),
	}

	collection := config.Ulbimongoconn.Collection("koleksi")
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan ke database: " + err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message":   "Koleksi berhasil disimpan.",
		"image_url": imageURL,
	})
}

// =============================================================
// 🟣 Fungsi Upload Gambar ke GitHub
// =============================================================
func uploadImageToGitHub(file *multipart.FileHeader, namaBenda string) (string, error) {
	githubToken := os.Getenv("GH_ACCESS_TOKEN") // Pastikan sudah di-set
	repoOwner := "ghaidafasya24"
	repoName := "images-koleksi-museum"
	filePath := fmt.Sprintf("koleksi/%d_%s.jpg", time.Now().Unix(), namaBenda)

	if githubToken == "" {
		return "", fmt.Errorf("GH_ACCESS_TOKEN belum diatur di environment variable")
	}

	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer f.Close()

	imageData, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %w", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)
	payload := map[string]string{
		"message": fmt.Sprintf("Upload image for %s", namaBenda),
		"content": encodedImage,
	}
	payloadBytes, _ := json.Marshal(payload)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

	req, _ := http.NewRequest("PUT", apiURL, bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal request ke GitHub API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	content, ok := result["content"].(map[string]interface{})
	if !ok || content["download_url"] == nil {
		return "", fmt.Errorf("tidak menemukan download_url dari GitHub response")
	}

	return content["download_url"].(string), nil
}

func GetAllKoleksi(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	col := db.Collection("koleksi") // nama koleksi MongoDB

	filter := bson.M{}
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error GetAllKoleksi:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal mengambil data koleksi",
			"error":   err.Error(),
		})
	}

	var koleksi []model.Koleksi
	err = cursor.All(context.TODO(), &koleksi)
	if err != nil {
		fmt.Println("Error decode:", err)
		return c.Status(500).JSON(fiber.Map{
			"message": "Gagal decode data koleksi",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil semua data koleksi",
		"total":   len(koleksi),
		"data":    koleksi,
	})
}

func GetKoleksiByID(c *fiber.Ctx) error {
	// Ambil parameter ID dari URL
	idParam := c.Params("id")

	// Konversi ID string menjadi ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID koleksi tidak valid",
		})
	}

	// Ambil koleksi koleksi
	collection := config.Ulbimongoconn.Collection("koleksi")

	// Query ke database
	var koleksi model.Koleksi
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&koleksi)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Koleksi tidak ditemukan",
		})
	}

	// Return hasil
	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil data koleksi",
		"data":    koleksi,
	})
}

func UpdateKoleksi(c *fiber.Ctx) error {
	id := c.Params("id")

	koleksiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID koleksi tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Ulbimongoconn.Collection("koleksi")

	// =========================
	// 🔹 Ambil data lama
	// =========================
	var existing model.Koleksi
	if err := collection.FindOne(ctx, bson.M{"_id": koleksiID}).Decode(&existing); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Koleksi tidak ditemukan"})
	}

	// =========================
	// 🔹 Form Value
	// =========================
	kategoriID := c.FormValue("kategori_id")
	gudangID := c.FormValue("gudang_id") // WAJIB
	rakID := c.FormValue("rak_id")       // OPSIONAL
	tahapID := c.FormValue("tahap_id")   // OPSIONAL
	hapusFoto := c.FormValue("hapus_foto")

	noReg := c.FormValue("no_reg")
	noInv := c.FormValue("no_inv")
	namaBenda := c.FormValue("nama_benda")
	asalKoleksi := c.FormValue("asal_koleksi")
	bahan := c.FormValue("bahan")
	tempatPerolehan := c.FormValue("tempat_perolehan")
	tanggalPerolehan := c.FormValue("tanggal_perolehan")
	deskripsi := c.FormValue("deskripsi")
	kondisi := c.FormValue("kondisi")

	// =========================
	// 🔹 KATEGORI
	// =========================
	kategori := existing.Kategori
	if kategoriID != "" {
		objID, err := primitive.ObjectIDFromHex(kategoriID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID kategori tidak valid"})
		}

		if err := config.Ulbimongoconn.
			Collection("kategori").
			FindOne(ctx, bson.M{"_id": objID}).
			Decode(&kategori); err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
		}
	}

	// =========================
	// 🔹 GUDANG (WAJIB)
	// =========================
	if gudangID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Gudang wajib diisi"})
	}

	objGudangID, _ := primitive.ObjectIDFromHex(gudangID)
	var gudang model.Gudang
	if err := config.Ulbimongoconn.
		Collection("gudang").
		FindOne(ctx, bson.M{"_id": objGudangID}).
		Decode(&gudang); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Gudang tidak ditemukan"})
	}

	// =========================
	// 🔹 RAK & TAHAP (OPSIONAL)
	// =========================
	var rak model.Rak
	if rakID != "" {
		objID, _ := primitive.ObjectIDFromHex(rakID)
		config.Ulbimongoconn.Collection("rak").
			FindOne(ctx, bson.M{"_id": objID}).Decode(&rak)
	}

	var tahap model.Tahap
	if tahapID != "" {
		objID, _ := primitive.ObjectIDFromHex(tahapID)
		config.Ulbimongoconn.Collection("tahap").
			FindOne(ctx, bson.M{"_id": objID}).Decode(&tahap)
	}

	// =========================
	// 🔹 TEMPAT PENYIMPANAN
	// =========================
	tempatPenyimpanan := model.TempatPenyimpanan{
		Gudang: gudang,
	}
	if rakID != "" {
		tempatPenyimpanan.Rak = rak
	}
	if tahapID != "" {
		tempatPenyimpanan.Tahap = tahap
	}

	// =========================
	// 🔹 FOTO LOGIC (INTI)
	// =========================
	setData := bson.M{
		"kategori":           kategori,
		"no_reg":             ifNotEmpty(noReg, existing.NoRegistrasi),
		"no_inv":             ifNotEmpty(noInv, existing.NoInventaris),
		"nama_benda":         ifNotEmpty(namaBenda, existing.NamaBenda),
		"asal_koleksi":       ifNotEmpty(asalKoleksi, existing.AsalKoleksi),
		"bahan":              ifNotEmpty(bahan, existing.Bahan),
		"tempat_perolehan":   ifNotEmpty(tempatPerolehan, existing.TempatPerolehan),
		"tanggal_perolehan":  ifNotEmpty(tanggalPerolehan, existing.TanggalPerolehan),
		"deskripsi":          ifNotEmpty(deskripsi, existing.Deskripsi),
		"kondisi":            ifNotEmpty(kondisi, existing.Kondisi),
		"tempat_penyimpanan": tempatPenyimpanan,
		"updated_at":         time.Now(),
	}

	unsetData := bson.M{}

	// ➕ Upload foto baru
	file, err := c.FormFile("foto")
	if err == nil && file != nil {
		imageURL, err := uploadImageToGitHub(file, namaBenda)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		setData["foto"] = imageURL
	}

	// ❌ Hapus foto
	if hapusFoto == "true" {
		unsetData["foto"] = ""
	}

	update := bson.M{"$set": setData}
	if len(unsetData) > 0 {
		update["$unset"] = unsetData
	}

	// =========================
	// 🔹 UPDATE DB
	// =========================
	if _, err := collection.UpdateOne(ctx, bson.M{"_id": koleksiID}, update); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update data"})
	}

	return c.JSON(fiber.Map{
		"message": "Koleksi berhasil diperbarui",
	})
}

// Fungsi helper agar field kosong tidak menimpa data lama
func ifNotEmpty(newValue, oldValue string) string {
	if newValue != "" {
		return newValue
	}
	return oldValue
}

// delete koleksi by ID
func DeleteKoleksiByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter URL
	idParam := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID koleksi tidak valid",
		})
	}

	// Koneksi ke database dan tentukan koleksi
	db := config.Ulbimongoconn
	col := db.Collection("koleksi")

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
		"message": fmt.Sprintf("Koleksi dengan ID %s berhasil dihapus", idParam),
	})
}
