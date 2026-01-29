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

// InsertKoleksi godoc
// @Summary      Insert Koleksi
// @Description  Menambahkan data koleksi museum baru, termasuk kategori, tempat penyimpanan, ukuran, foto, dan lain-lain
// @Tags         Data Koleksi
// @Accept       multipart/form-data
// @Produce      json
// @Param        no_reg             formData string true  "Nomor Registrasi"
// @Param        no_inv             formData string true  "Nomor Inventaris"
// @Param        nama_benda         formData string true  "Nama Benda"
// @Param        deskripsi          formData string false "Deskripsi Koleksi"
// @Param        kategori_id        formData string true  "ID Kategori"
// @Param        bahan              formData string false "Bahan Benda"
// @Param        tempat_perolehan   formData string false "Tempat Perolehan"
// @Param        tanggal_perolehan  formData string false "Tanggal Perolehan (format: DD-MM-YYYY)"
// @Param        panjang_keseluruhan formData string false "Panjang Keseluruhan (ukuran)"
// @Param        lebar              formData string false "Lebar (ukuran)"
// @Param        tebal              formData string false "Tebal (ukuran)"
// @Param        tinggi             formData string false "Tinggi (ukuran)"
// @Param        diameter           formData string false "Diameter (ukuran)"
// @Param        satuan             formData string false "Satuan ukuran keseluruhan (cm/m/dll)"
// @Param        berat              formData string false "Berat (ukuran)"
// @Param        satuan_berat       formData string false "Satuan berat (kg/g/dll)"
// @Param        gudang_id          formData string true  "ID Gudang"
// @Param        rak_id             formData string false "ID Rak"
// @Param        tahap_id           formData string false "ID Tahap"
// @Param        asal_koleksi       formData string false "Asal Koleksi"
// @Param        kondisi            formData string false "Kondisi Koleksi"
// @Param        foto               formData file   false "Upload foto koleksi"
// @Success      201 {object} map[string]interface{} "Koleksi berhasil disimpan"
// @Router       /koleksi [post]
// @Security     BearerAuth
func InsertKoleksi(c *fiber.Ctx) error {

	noReg := c.FormValue("no_reg")
	noInv := c.FormValue("no_inv")
	namaBenda := c.FormValue("nama_benda")
	tanggalPerolehan := c.FormValue("tanggal_perolehan")
	kategoriID := c.FormValue("kategori_id") // 🔹 ambil ID kategori, bukan nama
	bahan := c.FormValue("bahan")
	asalKoleksi := c.FormValue("asal_koleksi")
	tempatPerolehan := c.FormValue("tempat_perolehan")
	deskripsi := c.FormValue("deskripsi")
	Kondisi := c.FormValue("kondisi")

	// Tempat Penyimpanan
	gudangID := c.FormValue("gudang_id") // 🔹 ambil ID gudang, bukan nama
	rakID := c.FormValue("rak_id")       // 🔹 ambil ID rak, bukan nama
	catatan := c.FormValue("catatan")    // 🔹 ambil catatan untuk tempat penyimpanan
	tahapID := c.FormValue("tahap_id")   // 🔹 ambil ID tahap, bukan nama

	// Ukuran
	panjang := c.FormValue("panjang_keseluruhan")
	lebar := c.FormValue("lebar")
	tebal := c.FormValue("tebal")
	tinggi := c.FormValue("tinggi")
	diameter := c.FormValue("diameter")
	satuan := c.FormValue("satuan")

	berat := c.FormValue("berat")
	satuanBerat := c.FormValue("satuan_berat")

	// 🔹 cek apakah salah satu dimensi diisi
	adaDimensi := panjang != "" || lebar != "" || tebal != "" || tinggi != "" || diameter != ""

	// Validasi ukuran dan satuan
	if adaDimensi && satuan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Satuan wajib diisi jika salah satu dimensi ukuran diisi.",
		})
	}

	// Validasi harus mengisi dimensi jika mengisi satuan
	if !adaDimensi && satuan != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak boleh mengisi satuan tanpa mengisi dimensi ukuran.",
		})
	}

	//======================================================
	// Validasi berat dan satuan berat
	if berat != "" && satuanBerat == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Satuan berat wajib diisi jika berat diisi.",
		})
	}
	// Validasi harus mengisi satuan berat jika mengisi berat
	if berat == "" && satuanBerat != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak boleh mengisi satuan berat tanpa mengisi berat.",
		})
	}

	var ukuran *model.Ukuran

	if adaDimensi || berat != "" {
		ukuran = &model.Ukuran{
			PanjangKeseluruhan: panjang,
			Lebar:              lebar,
			Tebal:              tebal,
			Tinggi:             tinggi,
			Diameter:           diameter,
			Berat:              berat,
			Satuan:             satuan,
			SatuanBerat:        satuanBerat,
		}
	}

	if kategoriID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID Kategori tidak boleh kosong.",
		})
	}
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

		rakCollection := config.Ulbimongoconn.Collection("rak")
		err = rakCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&rak)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Data rak tidak ditemukan.",
			})
		}
	}

	// ======================================================
	// 🔹 TAHAP OPSIONAL
	// ======================================================
	var tahap model.Tahap
	if tahapID != "" {
		objID, err = primitive.ObjectIDFromHex(tahapID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ID tahap tidak valid.",
			})
		}

		tahapCollection := config.Ulbimongoconn.Collection("tahap")
		err = tahapCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&tahap)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Data tahap tidak ditemukan.",
			})
		}
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

	tempatPenyimpanan := model.TempatPenyimpanan{
		Catatan: catatan,
	}

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

	// ======================================================
	// VALIDASI UNIQUE no_reg
	// ======================================================
	count, err := collection.CountDocuments(ctx, bson.M{
		"no_reg": noReg,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengecek no registrasi.",
		})
	}
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No registrasi sudah digunakan.",
		})
	}

	// ======================================================
	// VALIDASI UNIQUE no_inv
	// ======================================================
	count, err = collection.CountDocuments(ctx, bson.M{
		"no_inv": noInv,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengecek no inventaris.",
		})
	}
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No inventaris sudah digunakan.",
		})
	}
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
	githubToken := os.Getenv("GH_ACCESS_TOKEN")
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

// GetAllKoleksi godoc
// @Summary      Get All Koleksi
// @Description  Mengambil semua data koleksi museum beserta kategori, tempat penyimpanan, dan ukuran
// @Tags         Data Koleksi
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /koleksi [get]
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

// GetKoleksiByID godoc
// @Summary      Get Koleksi By ID
// @Description  Mengambil satu data koleksi museum berdasarkan ID MongoDB
// @Tags         Data Koleksi
// @Produce      json
// @Param        id   path      string  true  "ID Koleksi"
// @Success      200  {object}  map[string]interface{}
// @Router       /koleksi/{id} [get]
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

// UpdateKoleksi godoc
// @Summary      Update Koleksi
// @Description  Memperbarui data koleksi museum berdasarkan ID. Semua field bersifat opsional, kecuali "gudang_id" wajib diisi. Jika foto diupload, akan mengganti foto lama.
// @Tags         Data Koleksi
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                 path     string true  "ID Koleksi"
// @Param        no_reg             formData string false "Nomor Registrasi"
// @Param        no_inv             formData string false "Nomor Inventaris"
// @Param        nama_benda         formData string false "Nama Benda"
// @Param        tanggal_perolehan  formData string false "Tanggal Perolehan (format: DD-MM-YYYY)"
// @Param        kategori_id        formData string false "ID Kategori"
// @Param        bahan              formData string false "Bahan Benda"
// @Param        panjang_keseluruhan formData string false "Panjang Keseluruhan (ukuran)"
// @Param        lebar              formData string false "Lebar (ukuran)"
// @Param        tebal              formData string false "Tebal (ukuran)"
// @Param        tinggi             formData string false "Tinggi (ukuran)"
// @Param        diameter           formData string false "Diameter (ukuran)"
// @Param        satuan             formData string false "Satuan ukuran keseluruhan (cm/m/dll)"
// @Param        berat              formData string false "Berat (ukuran)"
// @Param        satuan_berat       formData string false "Satuan berat (kg/g/dll)"
// @Param        asal_koleksi       formData string false "Asal Koleksi"
// @Param        tempat_perolehan   formData string false "Tempat Perolehan"
// @Param        deskripsi          formData string false "Deskripsi Koleksi"
// @Param        gudang_id          formData string true  "ID Gudang"
// @Param        rak_id             formData string false "ID Rak"
// @Param        tahap_id           formData string false "ID Tahap"
// @Param        kondisi            formData string false "Kondisi Koleksi"
// @Param        foto               formData file   false "Upload foto koleksi"
// @Success      200 {object} map[string]string "Koleksi berhasil diperbarui"
// @Router       /koleksi/{id} [put]
// @Security     BearerAuth
func UpdateKoleksi(c *fiber.Ctx) error {
	// =========================
	// VALIDASI ID PARAM
	// =========================
	id := c.Params("id")
	koleksiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID koleksi tidak valid"})
	}

	// =========================
	// CONTEXT
	// =========================
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Ulbimongoconn.Collection("koleksi")

	// =========================
	// AMBIL DATA LAMA
	// =========================
	var existing model.Koleksi
	if err := collection.FindOne(ctx, bson.M{"_id": koleksiID}).Decode(&existing); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Koleksi tidak ditemukan"})
	}

	// =========================
	// FORM DATA
	// =========================
	kategoriID := c.FormValue("kategori_id")
	gudangID := c.FormValue("gudang_id")

	noReg := c.FormValue("no_reg")
	noInv := c.FormValue("no_inv")
	namaBenda := c.FormValue("nama_benda")

	rakID := c.FormValue("rak_id")
	tahapID := c.FormValue("tahap_id")
	catatan := c.FormValue("catatan")

	asalKoleksi := c.FormValue("asal_koleksi")
	bahan := c.FormValue("bahan")
	tempatPerolehan := c.FormValue("tempat_perolehan")
	tanggalPerolehan := c.FormValue("tanggal_perolehan")
	deskripsi := c.FormValue("deskripsi")
	kondisi := c.FormValue("kondisi")

	// =========================
	// FORM DATA - UKURAN
	// =========================
	panjang := c.FormValue("panjang_keseluruhan")
	lebar := c.FormValue("lebar")
	tebal := c.FormValue("tebal")
	tinggi := c.FormValue("tinggi")
	diameter := c.FormValue("diameter")
	satuan := c.FormValue("satuan")

	berat := c.FormValue("berat")
	satuanBerat := c.FormValue("satuan_berat")

	adaDimensi := panjang != "" || lebar != "" || tebal != "" || tinggi != "" || diameter != ""

	// =========================
	// VALIDASI FIELD WAJIB
	// =========================
	if kategoriID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID Kategori tidak boleh kosong.",
		})
	}
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

	// =========================
	// VALIDASI UKURAN
	// =========================
	if adaDimensi && satuan == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Satuan wajib diisi jika salah satu dimensi ukuran diisi",
		})
	}

	if !adaDimensi && satuan != "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Tidak boleh mengisi satuan tanpa mengisi dimensi ukuran",
		})
	}

	if berat != "" && satuanBerat == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Satuan berat wajib diisi jika berat diisi",
		})
	}

	if berat == "" && satuanBerat != "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Tidak boleh mengisi satuan berat tanpa mengisi berat",
		})
	} // =========================
	// KATEGORI
	// =========================
	objKategoriID, err := primitive.ObjectIDFromHex(kategoriID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	var kategori model.Kategori
	if err := config.Ulbimongoconn.Collection("kategori").
		FindOne(ctx, bson.M{"_id": objKategoriID}).
		Decode(&kategori); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
	}

	// =========================
	// GUDANG
	// =========================
	objGudangID, err := primitive.ObjectIDFromHex(gudangID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID gudang tidak valid"})
	}

	var gudang model.Gudang
	if err := config.Ulbimongoconn.Collection("gudang").
		FindOne(ctx, bson.M{"_id": objGudangID}).
		Decode(&gudang); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Gudang tidak ditemukan"})
	}

	// =========================
	// PREPARE UPDATE
	// =========================
	setData := bson.M{}
	unsetData := bson.M{}

	// =========================
	// FIELD WAJIB → SELALU SET
	// =========================
	setData["kategori"] = kategori
	setData["no_reg"] = noReg
	setData["no_inv"] = noInv
	setData["nama_benda"] = namaBenda
	setData["tempat_penyimpanan.gudang"] = gudang

	// =========================
	// OPSIONAL: RAK
	// =========================
	if rakID != "" {
		objRakID, _ := primitive.ObjectIDFromHex(rakID)
		var rak model.Rak
		config.Ulbimongoconn.Collection("rak").
			FindOne(ctx, bson.M{"_id": objRakID}).
			Decode(&rak)
		setData["tempat_penyimpanan.rak"] = rak
	} else {
		unsetData["tempat_penyimpanan.rak"] = ""
	}

	// =========================
	// OPSIONAL: TAHAP
	// =========================
	if tahapID != "" {
		objTahapID, _ := primitive.ObjectIDFromHex(tahapID)
		var tahap model.Tahap
		config.Ulbimongoconn.Collection("tahap").
			FindOne(ctx, bson.M{"_id": objTahapID}).
			Decode(&tahap)
		setData["tempat_penyimpanan.tahap"] = tahap
	} else {
		unsetData["tempat_penyimpanan.tahap"] = ""
	}

	// =========================
	// OPSIONAL: CATATAN
	// =========================
	if catatan != "" {
		setData["tempat_penyimpanan.catatan"] = catatan
	} else {
		unsetData["tempat_penyimpanan.catatan"] = ""
	}

	// =========================
	// OPSIONAL FIELD LAIN
	// =========================
	handleOptional := func(key, value string) {
		if value != "" {
			setData[key] = value
		} else {
			unsetData[key] = ""
		}
	}

	handleOptional("asal_koleksi", asalKoleksi)
	handleOptional("bahan", bahan)
	handleOptional("tempat_perolehan", tempatPerolehan)
	handleOptional("tanggal_perolehan", tanggalPerolehan)
	handleOptional("deskripsi", deskripsi)
	handleOptional("kondisi", kondisi)

	// =========================
	// FOTO (INI PENTING 🔥)
	// =========================
	file, err := c.FormFile("foto")
	if err == nil && file != nil {
		imageURL, err := uploadImageToGitHub(file, namaBenda)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		setData["foto"] = imageURL
	} else {
		unsetData["foto"] = ""
	}

	// =========================
	// UKURAN (OPSIONAL)
	// =========================
	ukuran := model.Ukuran{}
	ukuran.PanjangKeseluruhan = c.FormValue("panjang_keseluruhan")
	ukuran.Lebar = c.FormValue("lebar")
	ukuran.Tebal = c.FormValue("tebal")
	ukuran.Tinggi = c.FormValue("tinggi")
	ukuran.Diameter = c.FormValue("diameter")
	ukuran.Berat = c.FormValue("berat")
	ukuran.Satuan = c.FormValue("satuan")
	ukuran.SatuanBerat = c.FormValue("satuan_berat")

	if ukuran != (model.Ukuran{}) {
		setData["ukuran"] = ukuran
	} else {
		unsetData["ukuran"] = ""
	}

	setData["updated_at"] = time.Now()

	// =========================
	// EXECUTE UPDATE
	// =========================
	update := bson.M{}
	if len(setData) > 0 {
		update["$set"] = setData
	}
	if len(unsetData) > 0 {
		update["$unset"] = unsetData
	}

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

// DeleteKoleksiByID godoc
// @Summary      Delete Koleksi by ID
// @Description  Menghapus data koleksi berdasarkan ID (wajib autentikasi JWT Bearer)
// @Tags         Data Koleksi
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID koleksi"
// @Router       /koleksi/{id} [delete]
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
