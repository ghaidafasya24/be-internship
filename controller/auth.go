package controller

import (
	"be-internship/config"
	"be-internship/model"
	"errors"
	"strings"

	// "be-internship/model"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register
// @Description Registrasi akun admin.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Payload Body [RAW]"
// @Success 200 {object} model.Users
// @Failure 400
// @Failure 500
// @Router       /users/register [post]
func Register(c *fiber.Ctx) error {
	// Context with timeout for MongoDB operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse request body into user model
	var user model.Users
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Validate required fields
	// if user.Username == "" || user.Password == "" || user.ConfirmPassword == "" || user.PhoneNumber == "" {
	if user.Username == "" || user.Password == "" || user.PhoneNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "All fields are required",
		})
	}

	// 🔹 VALIDASI FORMAT NOMOR TELEPON (HARUS 62)
	phone := user.PhoneNumber

	// Harus mulai dengan 62
	if !strings.HasPrefix(phone, "62") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format nomor telepon harus dimulai dengan 62",
		})
	}

	// Setelah 62 tidak boleh 0 → 6208... (SALAH)
	if len(phone) > 2 && phone[2] == '0' {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format nomor telepon tidak valid, gunakan format: 62xxxxxxxxxx (tanpa angka 0 setelah 62)",
		})
	}

	// Minimal panjang nomor telepon
	if len(phone) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nomor telepon terlalu pendek",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}
	user.Password = string(hashedPassword)

	// Check if username already exists
	usersCollection := config.Ulbimongoconn.Client().Database(config.DBUlbimongoinfo.DBName).Collection("users")
	var existingUser model.Users
	err = usersCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// Set additional fields
	user.ID = primitive.NewObjectID()

	// Set default role to "admin"
	user.Role = "admin"

	// Insert the new user into the database
	_, err = usersCollection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Respond with success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"status":  201,
		"user": fiber.Map{
			"_id":  user.ID,
			"role": user.Role,
		},
	})
}

var jwtKey = []byte("secret_key!234@!#$%")

// Claims struct untuk JWT
type Claims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

// Login godoc
// @Summary      Login
// @Description  Autentikasi user menggunakan username dan password, kemudian mengembalikan JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body      model.LoginRequest  true  "Data login user"
// @Success      200   {object}  model.Users
// @Failure 400
// @Failure 500
// @Router       /users/login [post]
func Login(c *fiber.Ctx) error {
	// Parse request body
	var loginData model.Users
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Cek apakah username ada di database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := config.Ulbimongoconn.Client().Database(config.DBUlbimongoinfo.DBName).Collection("users")
	var user model.Users
	err := usersCollection.FindOne(ctx, bson.M{"username": loginData.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to find user",
		})
	}

	// Verifikasi password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// // Generate JWT Token
	// expirationTime := time.Now().Add(24 * time.Hour) // Token berlaku selama 24 jam
	// claims := &Claims{
	// 	Username: user.Username,
	// 	RegisteredClaims: jwt.RegisteredClaims{
	// 		ExpiresAt: jwt.NewNumericDate(expirationTime),
	// 	},
	// }

	// Generate JWT Token dengan masa berlaku 30 menit
	expirationTime := time.Now().Add(60 * time.Hour)
	claims := &Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Kirim response dengan token JWT
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"status":  200,
		"role":    user.Role,
		"token":   tokenString,
		"expires": expirationTime,
	})
}

// ValidateToken memvalidasi token JWT
func ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

// JWTAuth middleware untuk memverifikasi token di Fiber
func JWTAuth(c *fiber.Ctx) error {

	// Ambil header Authorization
	bearerToken := c.Get("Authorization")
	if bearerToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "token tidak ditemukan",
		})
	}

	// Format harus: "Bearer <token>"
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "token tidak valid (format salah)",
		})
	}

	tokenString := tokenParts[1]

	// Parse token dan ambil claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	// Jika token rusak / signature salah
	if err != nil {
		// CEK APAKAH EXPIRED
		if errors.Is(err, jwt.ErrTokenExpired) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "token expired",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "token tidak valid",
		})
	}

	// Jika token tidak valid (false)
	if !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "token tidak valid",
		})
	}

	// Jika valid → lanjutkan handler berikutnya
	return c.Next()
}

// Get All Users godoc
// @Summary      Get All Users
// @Description  Endpoint untuk mengambil seluruh data admin yang tersimpan di sistem
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200 {object} model.GetAllUsersResponse
// @Failure 400
// @Failure 500
// @Router       /users [get]
func GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ambil collection users
	usersCollection := config.Ulbimongoconn.Client().Database(config.DBUlbimongoinfo.DBName).Collection("users")

	// Query semua data
	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data users",
		})
	}
	defer cursor.Close(ctx)

	// Decode hasil cursor
	var users []model.Users
	for cursor.Next(ctx) {
		var user model.Users
		if err := cursor.Decode(&user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal decode data user",
			})
		}

		// Hapus password dari response
		user.Password = ""
		// Jika tidak ingin menampilkan nomor HP:
		// user.PhoneNumber = ""

		users = append(users, user)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Berhasil mengambil semua users",
		"total":   len(users),
		"data":    users,
	})
}

// GetUserByUsername godoc
// @Summary      Get User by Username
// @Description  Mengambil data detail user berdasarkan username tertentu
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  model.GetAllUsersResponse
// @Failure 400
// @Failure 500
// @Router       /users/username/{username} [get]
func GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username wajib diisi",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := config.Ulbimongoconn.Client().
		Database(config.DBUlbimongoinfo.DBName).
		Collection("users")

	var user model.Users
	err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data user",
		})
	}

	// Hapus password sebelum dikirim
	user.Password = ""

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User ditemukan",
		"data":    user,
	})
}

// GetUserByID godoc
// @Summary      Get User by ID
// @Description  Mengambil data user berdasarkan ID MongoDB
// @Tags         Users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User berhasil ditampilkan"
// @Router       /users/{id} [get]
func GetUserByID(c *fiber.Ctx) error {
	// Ambil ID dari URL, trim spasi & quotes
	idParam := strings.TrimSpace(c.Params("id"))
	idParam = strings.Trim(idParam, `"`) // hapus quotes jika Swagger menambahkan

	// Validasi ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID user tidak valid",
		})
	}

	// Ambil collection users
	collection := config.Ulbimongoconn.Collection("users")

	var user model.Users
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data user",
		})
	}

	// Jangan kirim password
	user.Password = ""

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User berhasil ditampilkan",
		"data":    user,
	})
}

// UpdateUserByID godoc
// @Summary      Update User
// @Description  Memperbarui data user berdasarkan ID (wajib autentikasi JWT Bearer)
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id            path      string  true   "ID user"
// @Param        username      formData  string  false  "Username (huruf kecil semua)"
// @Param        phone_number  formData  string  false  "Nomor telepon format 62xxxxxxxx"
// @Param        password      formData  string  false  "Password baru"
// @Router       /users/{id} [put]
func UpdateUserByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID user wajib diisi",
		})
	}

	// Convert ID
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID user tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := config.
		Ulbimongoconn.Client().
		Database(config.DBUlbimongoinfo.DBName).
		Collection("users")

	// Cek user
	var existingUser model.Users
	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal mengambil data user",
		})
	}

	// Ambil data dari form-data
	username := c.FormValue("username")
	phone := c.FormValue("phone_number")
	password := c.FormValue("password")
	role := c.FormValue("role")

	update := bson.M{}

	// ----------------------------
	// USERNAME → wajib lowercase
	// ----------------------------
	if username != "" {
		lower := strings.ToLower(username)
		if username != lower {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username harus huruf kecil semua",
			})
		}
		update["username"] = lower
	}

	// ----------------------------
	// PHONE NUMBER → wajib 62xxx
	// ----------------------------
	if phone != "" {
		if !strings.HasPrefix(phone, "62") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Format nomor telepon harus dimulai dengan 62",
			})
		}

		if len(phone) > 2 && phone[2] == '0' {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Gunakan format 62xxxxxxxx (tanpa 0 setelah 62)",
			})
		}

		if len(phone) < 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Nomor telepon terlalu pendek",
			})
		}

		update["phone_number"] = phone
	}

	// ----------------------------
	// PASSWORD → langsung hash saja
	// ----------------------------
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Gagal mengenkripsi password",
			})
		}
		update["password"] = string(hashedPassword)
	}

	// ROLE (opsional)
	if role != "" {
		update["role"] = role
	}

	// Jika tidak ada yang dikirim
	if len(update) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tidak ada data yang diupdate",
		})
	}

	// Eksekusi update
	res, err := usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal update user",
		})
	}

	// Pastikan ada perubahan
	if res.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User tidak ditemukan",
		})
	}

	if res.ModifiedCount == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tidak ada perubahan data",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User berhasil diupdate",
		"id":      idParam,
	})
}

// DeleteUserByID godoc
// @Summary      Delete User
// @Description  Menghapus data user berdasarkan ID (wajib autentikasi JWT Bearer)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path   string  true  "ID user"
// @Success      200  {object}  map[string]interface{}  "User berhasil dihapus"
// @Router       /users/{id} [delete]
func DeleteUserByID(c *fiber.Ctx) error {
	// Ambil ID dari parameter URL
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID user wajib diisi",
		})
	}

	// Convert string ID ke ObjectID Mongo
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID user tidak valid",
		})
	}

	// Koneksi ke DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := config.
		Ulbimongoconn.Client().Database(config.DBUlbimongoinfo.DBName).
		Collection("users")

	// Cek apakah user ada
	var existingUser model.Users
	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User tidak ditemukan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal mengambil data user",
		})
	}

	// Hapus user
	_, err = usersCollection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal menghapus user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User berhasil dihapus",
		"id":      idParam,
	})
}
