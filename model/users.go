package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Users struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Role        string             `json:"role,omitempty" bson:"role,omitempty"`
	Username    string             `json:"username,omitempty" bson:"username,omitempty" gorm:"unique;not null"`
	PhoneNumber string             `json:"phone_number,omitempty" bson:"phone_number,omitempty" gorm:"unique;not null"`
	Password    string             `json:"password,omitempty" bson:"password,omitempty"`
	// ConfirmPassword string             `json:"confirm_password,omitempty" bson:"confirm_password,omitempty"`
	// ResetOTP       string    `json:"reset_otp,omitempty" bson:"reset_otp,omitempty"`
	// ResetOTPExpire time.Time `json:"reset_otp_expire,omitempty" bson:"reset_otp_expire,omitempty"`
}
