package models

import (
	"database/sql"
	"time"
)

// ==================== USER MODEL ====================
type User struct {
	ID           int            `json:"id"`
	Username     string         `json:"username"`
	Password     string         `json:"-"` // Hidden from JSON
	Fullname     string         `json:"fullname"`
	Email        string         `json:"email"`
	Role         string         `json:"role"`
	ProfileImage sql.NullString `json:"profile_image"`
	Avatar       string         `json:"avatar"` // For backward compatibility
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastLogin    *time.Time     `json:"last_login,omitempty"`
}

// ==================== AUTH REQUEST ====================
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"required,oneof=superadmin head produksi"`
	Avatar   string `json:"avatar"`
}

// ==================== PROFILE REQUEST ====================
type UpdateProfileInput struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Fullname string `json:"fullname" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

// ==================== USER RESPONSE ====================
type UserResponse struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Fullname     string     `json:"fullname"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	ProfileImage *string    `json:"profile_image"`
	Avatar       string     `json:"avatar"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

// ==================== USER FILTER ====================
type UserFilter struct {
	Search string `json:"search"`
	Role   string `json:"role"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// ==================== HELPER FUNCTIONS ====================
// ToUserResponse converts User to UserResponse
func (u *User) ToUserResponse() UserResponse {
	var profileImage *string
	if u.ProfileImage.Valid {
		profileImage = &u.ProfileImage.String
	}

	return UserResponse{
		ID:           u.ID,
		Username:     u.Username,
		Fullname:     u.Fullname,
		Email:        u.Email,
		Role:         u.Role,
		ProfileImage: profileImage,
		Avatar:       u.Avatar,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		LastLogin:    u.LastLogin,
	}
}