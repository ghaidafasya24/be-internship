package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Users struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" example:"12345678"`
	Role        string             `json:"role,omitempty" bson:"role,omitempty" example:"admin"`
	Username    string             `json:"username,omitempty" bson:"username,omitempty" gorm:"unique;not null" example:"ghaida"`
	PhoneNumber string             `json:"phone_number,omitempty" bson:"phone_number,omitempty" gorm:"unique;not null" example:"6281234567890"`
	Password    string             `json:"password,omitempty" bson:"password,omitempty" example:"admin12345" swaggerignore:"true"`
}

type RegisterRequest struct {
	Username    string `json:"username,omitempty" bson:"username,omitempty" gorm:"unique;not null" example:"ghaida"`
	PhoneNumber string `json:"phone_number,omitempty" bson:"phone_number,omitempty" gorm:"unique;not null" example:"6281234567890"`
	Password    string `json:"password,omitempty" bson:"password,omitempty" example:"admin12345"`
}

type LoginRequest struct {
	Username string `json:"username" example:"ghaida"`
	Password string `json:"password" example:"admin12345"`
}

type GetAllUsersResponse struct {
	Message string  `json:"message" example:"Berhasil mengambil semua data users"`
	Total   int     `json:"total" example:"1"`
	Data    []Users `json:"data"`
}
