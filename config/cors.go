package config

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2/middleware/cors"

)

var origins = []string{
	"http://127.0.0.1:5500",
	"http://127.0.0.1:5501",
	"http://127.0.0.1:8080",
	"https://ghaidafasya24.github.io",
	"http://127.0.0.1:44857",
	"http://127.0.0.1:3000",
}

var Internalhost string = os.Getenv("INTERNALHOST") + ":" + os.Getenv("PORT")

var Cors = cors.Config{
	AllowOrigins:     strings.Join(origins[:], ","),
	AllowMethods:     "GET,HEAD,OPTIONS,POST,PUT,DELETE",
	AllowHeaders:     "Origin,Login,Content-Type,Authorization",
	ExposeHeaders:    "Content-Length",
	AllowCredentials: true,
}
