package model

import "time"

// REGISTER
// RegisterRequest untuk request registrasi user baru
type RegisterRequest struct {
	Username    string `json:"username,omitempty" bson:"username,omitempty" gorm:"unique;not null" example:"ghaida"`
	PhoneNumber string `json:"phone_number,omitempty" bson:"phone_number,omitempty" gorm:"unique;not null" example:"6281234567890"`
	Password    string `json:"password,omitempty" bson:"password,omitempty" example:"admin12345"`
}

// RegisterResponse untuk response sukses registrasi
type RegisterResponse struct {
	Message string `json:"message" example:"User registered successfully"`
	Status  int    `json:"status" example:"201"`
	User    struct {
		ID   string `json:"_id" example:"696ef88677f450e9430a144e"`
		Role string `json:"role" example:"admin"`
	} `json:"user"`
}

// ErrorResponseRegister untuk response error registrasi
type ErrorResponseRegister struct {
	Error string `json:"error" example:"Username already exists"`
}

// LOGIN
// LoginRequest untuk request login user
type LoginRequest struct {
	Username string `json:"username" example:"ghaida"`
	Password string `json:"password" example:"admin12345"`
}

// LoginResponse untuk response sukses login
type LoginResponse struct {
	Message string    `json:"message" example:"Login successful"`
	Status  int       `json:"status" example:"200"`
	Role    string    `json:"role" example:"admin"`
	Token   string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Expires time.Time `json:"expires" example:"2026-01-22T15:11:51.917322007Z"`
}

// ErrorResponse untuk response error login
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}

// USERS
// Response sukses get user by username
type GetUserByUsernameResponse struct {
	Message string `json:"message" example:"User ditemukan"`
	Data    Users  `json:"data"`
}

// Response error user tidak ditemukan
type ErrorResponseUsername struct {
	Error string `json:"error" example:"User tidak ditemukan"`
}

// Response sukses get all users
type GetAllUsersResponse struct {
	Message string  `json:"message" example:"Berhasil mengambil semua data users"`
	Total   int     `json:"total" example:"1"`
	Data    []Users `json:"data"`
}

// RakResponseItem untuk item rak
type RakResponseItem struct {
	ID      string `json:"id" example:"693a3a7a416cd8d592b5058e"`
	NamaRak string `json:"nama_rak" example:"Rak 2"`
}

// TahapResponseItem untuk item tahap
type TahapResponseItem struct {
	ID        string `json:"id" example:"693a3a7a416cd8d59235fsa"`
	NamaTahap string `json:"nama_tahap" example:"Tahap 2"`
}

// GetAllRakResponse untuk response Get All Rak
type GetAllRakResponse struct {
	Message string            `json:"message" example:"Berhasil mengambil semua data rak"`
	Data    []RakResponseItem `json:"data"`
	Total   int               `json:"total" example:"12"`
}

// GetAllRakResponse untuk response Get All Rak
type GetAllTahapResponse struct {
	Message string              `json:"message" example:"Berhasil mengambil semua data tahap"`
	Data    []TahapResponseItem `json:"data"`
	Total   int                 `json:"total" example:"12"`
}
