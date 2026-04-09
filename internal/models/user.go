package models

import "time"

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleOperator UserRole = "operator"
	RoleOwner    UserRole = "owner"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID            string    `json:"id" db:"id"`
	FullName      string    `json:"full_name" db:"full_name"`
	Email         string    `json:"email" db:"email"`
	Mobile        *string   `json:"mobile,omitempty" db:"mobile"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	Role          UserRole  `json:"role" db:"role"`
	WalletBalance float64   `json:"wallet_balance" db:"wallet_balance"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}