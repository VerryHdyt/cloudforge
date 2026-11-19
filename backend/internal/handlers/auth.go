package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/VerryHdyt/cloudforge/backend/internal/models"
	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB    *gorm.DB
	Audit *services.AuditService
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// recordAudit records authentication activity without interrupting
// the main authentication process if audit logging fails.
func (h *AuthHandler) recordAudit(
	c *gin.Context,
	userID *uint,
	action string,
	resourceID string,
) {
	if h.Audit == nil {
		log.Println("Audit service is not configured")
		return
	}

	if err := h.Audit.LogFromContext(
		c,
		userID,
		action,
		"authentication",
		resourceID,
	); err != nil {
		log.Printf("Failed to record audit log: %v", err)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name cannot be empty"})
		return
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "viewer",
	}

	if err := h.DB.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	userID := user.ID
	h.recordAudit(c, &userID, "register", strconv.FormatUint(uint64(user.ID), 10))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"data":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User

	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			h.recordAudit(c, nil, "login_failed", "")

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	if !services.CheckPassword(req.Password, user.PasswordHash) {
		h.recordAudit(c, nil, "login_failed", "")

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := services.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	userID := user.ID
	h.recordAudit(c, &userID, "login", strconv.FormatUint(uint64(user.ID), 10))

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successful",
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user":         user,
	})
}
