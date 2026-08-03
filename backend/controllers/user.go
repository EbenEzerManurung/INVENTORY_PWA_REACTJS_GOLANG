package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"inventory-backend/config"
	"inventory-backend/models"
	"inventory-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ==================== GET CURRENT USER PROFILE ====================
func GetCurrentUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	var user models.User
	query := `
		SELECT id, username, fullname, email, role, profile_image, 
			   created_at, updated_at, last_login
		FROM users 
		WHERE id = ?
	`
	err := config.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Fullname, &user.Email,
		&user.Role, &user.ProfileImage, &user.CreatedAt,
		&user.UpdatedAt, &user.LastLogin,
	)
	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user profile", err.Error())
		return
	}

	// Set Avatar for backward compatibility
	if user.ProfileImage.Valid {
		user.Avatar = user.ProfileImage.String
	}

	utils.Success(c, "Profile retrieved successfully", user.ToUserResponse())
}

// ==================== UPDATE USER PROFILE ====================
func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	var input models.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Check if username already exists
	if input.Username != "" {
		var count int
		err := config.DB.QueryRow(
			"SELECT COUNT(*) FROM users WHERE username = ? AND id != ?",
			input.Username, userID,
		).Scan(&count)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Database error", err.Error())
			return
		}
		if count > 0 {
			utils.Error(c, http.StatusBadRequest, "Username already taken", "")
			return
		}
	}

	// Check if email already exists
	if input.Email != "" {
		var count int
		err := config.DB.QueryRow(
			"SELECT COUNT(*) FROM users WHERE email = ? AND id != ?",
			input.Email, userID,
		).Scan(&count)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Database error", err.Error())
			return
		}
		if count > 0 {
			utils.Error(c, http.StatusBadRequest, "Email already taken", "")
			return
		}
	}

	// Build update query
	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if input.Username != "" {
		updates = append(updates, "username = ?")
		args = append(args, input.Username)
	}
	if input.Fullname != "" {
		updates = append(updates, "fullname = ?")
		args = append(args, input.Fullname)
	}
	if input.Email != "" {
		updates = append(updates, "email = ?")
		args = append(args, input.Email)
	}

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "No fields to update", "")
		return
	}

	updates = append(updates, "updated_at = NOW()")
	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, userID)

	_, err := config.DB.Exec(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	// Get updated user
	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, username, fullname, email, role, profile_image, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Fullname, &user.Email, &user.Role, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to fetch updated user", err.Error())
		return
	}

	utils.Success(c, "Profile updated successfully", user.ToUserResponse())
}

// ==================== CHANGE PASSWORD ====================
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	var input models.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Get current password
	var currentPassword string
	err := config.DB.QueryRow(
		"SELECT password FROM users WHERE id = ?",
		userID,
	).Scan(&currentPassword)
	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(input.CurrentPassword))
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Current password is incorrect", "")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	// Update password
	_, err = config.DB.Exec(
		"UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?",
		hashedPassword, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	utils.Success(c, "Password updated successfully", nil)
}

// ==================== UPLOAD AVATAR ====================
func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	// Get file from request
	file, err := c.FormFile("avatar")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "No file uploaded", err.Error())
		return
	}

	// Validate file size (max 2MB)
	if file.Size > 2*1024*1024 {
		utils.Error(c, http.StatusBadRequest, "File size must be less than 2MB", "")
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		utils.Error(c, http.StatusBadRequest, "Only image files are allowed (jpeg, png, gif, webp)", "")
		return
	}

	// Create upload directory if not exists
	uploadDir := "uploads/profiles"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create upload directory", err.Error())
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%d%s", userID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to save file", err.Error())
		return
	}

	// Delete old avatar if exists
	var oldAvatar sql.NullString
	err = config.DB.QueryRow(
		"SELECT profile_image FROM users WHERE id = ?",
		userID,
	).Scan(&oldAvatar)
	if err == nil && oldAvatar.Valid && oldAvatar.String != "" {
		oldPath := strings.TrimPrefix(oldAvatar.String, "/uploads/profiles/")
		if oldPath != "" {
			oldFullPath := filepath.Join(uploadDir, oldPath)
			os.Remove(oldFullPath)
		}
	}

	// Update database
	avatarURL := "/uploads/profiles/" + filename
	_, err = config.DB.Exec(
		"UPDATE users SET profile_image = ?, updated_at = NOW() WHERE id = ?",
		avatarURL, userID,
	)
	if err != nil {
		os.Remove(filePath)
		utils.Error(c, http.StatusInternalServerError, "Failed to update avatar", err.Error())
		return
	}

	// ✅ Ambil data user terbaru
	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, username, fullname, email, role, profile_image, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(
		&user.ID, &user.Username, &user.Fullname, &user.Email,
		&user.Role, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt,
	)

	// ✅ Kirim response dengan user data
	if err == nil {
		userResponse := user.ToUserResponse()
		utils.Success(c, "Avatar uploaded successfully", gin.H{
			"avatar_url": avatarURL,
			"user":       userResponse,
		})
	} else {
		utils.Success(c, "Avatar uploaded successfully", gin.H{
			"avatar_url": avatarURL,
		})
	}
}

// ==================== REMOVE AVATAR ====================
func RemoveAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	// Get current avatar
	var avatar sql.NullString
	err := config.DB.QueryRow(
		"SELECT profile_image FROM users WHERE id = ?",
		userID,
	).Scan(&avatar)
	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Delete file
	if avatar.Valid && avatar.String != "" {
		avatarPath := strings.TrimPrefix(avatar.String, "/uploads/profiles/")
		if avatarPath != "" {
			os.Remove(filepath.Join("uploads/profiles", avatarPath))
		}
	}

	// Update database
	_, err = config.DB.Exec(
		"UPDATE users SET profile_image = NULL, updated_at = NOW() WHERE id = ?",
		userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to remove avatar", err.Error())
		return
	}

	utils.Success(c, "Avatar removed successfully", nil)
}

// ==================== GET ALL USERS ====================
func GetUsers(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var whereClause string
	var args []interface{}

	if search != "" {
		whereClause = "WHERE username LIKE ? OR fullname LIKE ? OR email LIKE ? OR role LIKE ?"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	query := "SELECT id, username, fullname, email, role, profile_image, created_at, updated_at FROM users " +
		whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Fullname,
			&user.Email,
			&user.Role,
			&user.ProfileImage,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to scan users", err.Error())
			return
		}
		users = append(users, user.ToUserResponse())
	}

	utils.PaginatedResponse(c, users, total, page, limit)
}

// ==================== GET USER BY ID ====================
func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var user models.User
	err = config.DB.QueryRow(`
		SELECT id, username, fullname, email, role, profile_image, created_at, updated_at 
		FROM users WHERE id = ?
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.Fullname,
		&user.Email,
		&user.Role,
		&user.ProfileImage,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	utils.Success(c, "User retrieved successfully", user.ToUserResponse())
}

// ==================== CREATE USER ====================
func CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Validate role
	validRoles := map[string]bool{"superadmin": true, "head": true, "produksi": true}
	if !validRoles[req.Role] {
		utils.Error(c, http.StatusBadRequest, "Invalid role. Must be: superadmin, head, or produksi", "")
		return
	}

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

	// Check email if provided
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	var profileImage interface{}
	if req.Avatar != "" {
		profileImage = req.Avatar
	} else {
		profileImage = nil
	}

	var email interface{}
	if req.Email != "" {
		email = req.Email
	} else {
		email = nil
	}

	result, err := config.DB.Exec(`
		INSERT INTO users (username, password, fullname, email, role, profile_image) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Username, hashedPassword, req.Fullname, email, req.Role, profileImage)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, "User created successfully", gin.H{"id": id})
}

// ==================== UPDATE USER ====================
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Fullname string `json:"fullname"`
		Role     string `json:"role"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Check if user exists
	var exists bool
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to check user", err.Error())
		return
	}
	if !exists {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	// Validate role if provided
	if req.Role != "" {
		validRoles := map[string]bool{"superadmin": true, "head": true, "produksi": true}
		if !validRoles[req.Role] {
			utils.Error(c, http.StatusBadRequest, "Invalid role. Must be: superadmin, head, or produksi", "")
			return
		}
	}

	// Check email if provided and changed
	if req.Email != "" {
		var count int
		err = config.DB.QueryRow(
			"SELECT COUNT(*) FROM users WHERE email = ? AND id != ?",
			req.Email, id,
		).Scan(&count)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Database error", err.Error())
			return
		}
		if count > 0 {
			utils.Error(c, http.StatusBadRequest, "Email already taken", "")
			return
		}
	}

	// Build update query
	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if req.Fullname != "" {
		updates = append(updates, "fullname = ?")
		args = append(args, req.Fullname)
	}
	if req.Role != "" {
		updates = append(updates, "role = ?")
		args = append(args, req.Role)
	}
	if req.Email != "" {
		updates = append(updates, "email = ?")
		args = append(args, req.Email)
	}
	if req.Avatar != "" {
		updates = append(updates, "profile_image = ?")
		args = append(args, req.Avatar)
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
			return
		}
		updates = append(updates, "password = ?")
		args = append(args, hashedPassword)
	}

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "No fields to update", "")
		return
	}

	updates = append(updates, "updated_at = NOW()")
	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err = config.DB.Exec(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	utils.Success(c, "User updated successfully", nil)
}

// ==================== DELETE USER ====================
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var role string
	err = config.DB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to check user", err.Error())
		return
	}

	// Prevent deleting last superadmin
	if role == "superadmin" {
		var count int
		err = config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'superadmin'").Scan(&count)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to count superadmin", err.Error())
			return
		}
		if count <= 1 {
			utils.Error(c, http.StatusBadRequest, "Cannot delete the last superadmin user", "")
			return
		}
	}

	// Delete user's avatar if exists
	var avatar sql.NullString
	err = config.DB.QueryRow("SELECT profile_image FROM users WHERE id = ?", id).Scan(&avatar)
	if err == nil && avatar.Valid && avatar.String != "" {
		avatarPath := strings.TrimPrefix(avatar.String, "/uploads/profiles/")
		if avatarPath != "" {
			os.Remove(filepath.Join("uploads/profiles", avatarPath))
		}
	}

	// Delete user (cascade will handle related records if set)
	_, err = config.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	utils.Success(c, "User deleted successfully", nil)
}

// ==================== UPDATE USER ROLE ====================
func UpdateUserRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=superadmin head produksi"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var exists bool
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to check user", err.Error())
		return
	}
	if !exists {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	// Prevent changing last superadmin role
	var currentRole string
	err = config.DB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&currentRole)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get user role", err.Error())
		return
	}

	if currentRole == "superadmin" && req.Role != "superadmin" {
		var count int
		err = config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'superadmin'").Scan(&count)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to count superadmin", err.Error())
			return
		}
		if count <= 1 {
			utils.Error(c, http.StatusBadRequest, "Cannot change the role of the last superadmin", "")
			return
		}
	}

	_, err = config.DB.Exec(
		"UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?",
		req.Role, id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update user role", err.Error())
		return
	}

	utils.Success(c, "User role updated successfully", nil)
}

// ==================== RESET USER PASSWORD ====================
func ResetUserPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var exists bool
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to check user", err.Error())
		return
	}
	if !exists {
		utils.Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	_, err = config.DB.Exec(
		"UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?",
		hashedPassword, id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to reset password", err.Error())
		return
	}

	utils.Success(c, "Password reset successfully", nil)
}