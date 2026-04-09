package auth

type RegisterRequest struct {
	FullName string  `json:"full_name" binding:"required,min=2,max=120"`
	Email    string  `json:"email" binding:"required,email"`
	Mobile   *string `json:"mobile,omitempty"`
	Password string  `json:"password" binding:"required,min=6,max=72"`
	Role     string  `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserPayload `json:"user"`
}

type UserPayload struct {
	ID            string   `json:"id"`
	FullName      string   `json:"full_name"`
	Email         string   `json:"email"`
	Mobile        *string  `json:"mobile,omitempty"`
	Role          string   `json:"role"`
	WalletBalance float64  `json:"wallet_balance"`
	IsActive      bool     `json:"is_active"`
}