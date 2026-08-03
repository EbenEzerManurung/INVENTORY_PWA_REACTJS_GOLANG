package controllers

import (
	"net/http"
	"time"

	"inventory-backend/config"
	"inventory-backend/models"
	"inventory-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var user models.User
	var hashedPassword string
	query := `
		SELECT id, username, password, fullname, role, profile_image, email
		FROM users 
		WHERE username = ? OR email = ?
	`
	err := config.DB.QueryRow(query, req.Username, req.Username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.Fullname,
		&user.Role, &user.ProfileImage, &user.Email,
	)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid username or password", "")
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid username or password", "")
		return
	}

	// Update last login
	now := time.Now()
	_, err = config.DB.Exec(
		"UPDATE users SET last_login = ? WHERE id = ?",
		now, user.ID,
	)
	if err != nil {
		// Log error but don't fail login
		println("Failed to update last_login:", err.Error())
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate refresh token", err.Error())
		return
	}

	// Set user response
	userResponse := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Fullname:  user.Fullname,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.ProfileImage.Valid {
		userResponse.ProfileImage = &user.ProfileImage.String
	}

	utils.Success(c, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          userResponse,
	})
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Check if username exists
	var exists bool
	err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}
	if exists {
		utils.Error(c, http.StatusBadRequest, "Username already exists", "")
		return
	}

	// Check if email exists (if provided)
	if req.Email != "" {
		err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Database error", err.Error())
			return
		}
		if exists {
			utils.Error(c, http.StatusBadRequest, "Email already exists", "")
			return
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	// Insert user
	var email interface{}
	if req.Email != "" {
		email = req.Email
	} else {
		email = nil
	}

	result, err := config.DB.Exec(`
		INSERT INTO users (username, password, fullname, email, role) 
		VALUES (?, ?, ?, ?, ?)
	`, req.Username, hashedPassword, req.Fullname, email, req.Role)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, "User registered successfully", gin.H{"id": id})
}

func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid refresh token", err.Error())
		return
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	utils.Success(c, "Token refreshed successfully", gin.H{
		"access_token": newAccessToken,
	})
}

func Logout(c *gin.Context) {
	utils.Success(c, "Logged out successfully", nil)
}

// ==================== GET CURRENT USER ====================
func GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	var user models.User
	query := `
		SELECT id, username, fullname, email, role, profile_image, created_at, updated_at
		FROM users WHERE id = ?
	`
	err := config.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Fullname, &user.Email,
		&user.Role, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	userResponse := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Fullname:  user.Fullname,
		Email:     user.Email,
		Role:      user.Role,
		ProfileImage: func() *string {
			if user.ProfileImage.Valid {
				return &user.ProfileImage.String
			}
			return nil
		}(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.Success(c, "User retrieved successfully", userResponse)
}