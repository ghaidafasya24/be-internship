package controller

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("secret_key!234@!#$%")

// Claims struct untuk JWT
type Claims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
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
