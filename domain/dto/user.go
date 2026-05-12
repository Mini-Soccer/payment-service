package dto

import "github.com/google/uuid"

type User struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	PhoneNumber string    `json:"phoneNumber"`
}
