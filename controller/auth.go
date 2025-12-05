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

// REGISTER
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

// LOGIN
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
	expirationTime := time.Now().Add(30 * time.Minute)
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

// GET ALL USERS
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

// GET USER BY USERNAME
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

// DELETE USER BY ID
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
		Ulbimongoconn.
		Client().
		Database(config.DBUlbimongoinfo.DBName).
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
