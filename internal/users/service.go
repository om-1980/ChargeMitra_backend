package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
	ID            string  `json:"id"`
	FullName      string  `json:"full_name"`
	Email         string  `json:"email"`
	Mobile        *string `json:"mobile,omitempty"`
	Role          string  `json:"role"`
	WalletBalance float64 `json:"wallet_balance"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetProfile(userID string) (*UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile UserProfile
	var createdAt time.Time
	var updatedAt time.Time

	err := s.db.QueryRow(ctx, `
		SELECT id, full_name, email, mobile, role, wallet_balance, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&profile.ID,
		&profile.FullName,
		&profile.Email,
		&profile.Mobile,
		&profile.Role,
		&profile.WalletBalance,
		&profile.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, errors.New("user not found")
	}

	profile.CreatedAt = createdAt.Format(time.RFC3339)
	profile.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &profile, nil
}