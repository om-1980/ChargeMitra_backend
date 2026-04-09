package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db        *pgxpool.Pool
	jwtSecret string
}

func NewService(db *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "customer"
	}

	switch role {
	case "customer", "operator", "owner", "admin":
	default:
		return nil, errors.New("invalid role")
	}

	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE email = $1
		)
	`, strings.ToLower(req.Email)).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user UserPayload
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (full_name, email, mobile, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, full_name, email, mobile, role, wallet_balance, is_active
	`,
		req.FullName,
		strings.ToLower(req.Email),
		req.Mobile,
		string(passwordHash),
		role,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Mobile,
		&user.Role,
		&user.WalletBalance,
		&user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	token, err := GenerateToken(s.jwtSecret, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		user         UserPayload
		passwordHash string
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, full_name, email, mobile, password_hash, role, wallet_balance, is_active
		FROM users
		WHERE email = $1
	`, strings.ToLower(req.Email)).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Mobile,
		&passwordHash,
		&user.Role,
		&user.WalletBalance,
		&user.IsActive,
	)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := GenerateToken(s.jwtSecret, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}