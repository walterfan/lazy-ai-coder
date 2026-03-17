package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/auth"
	"github.com/walterfan/lazy-ai-coder/pkg/authz"
)

// ProfileHandlers handles user profile endpoints
type ProfileHandlers struct {
	userService     *auth.OAuthUserService
	passwordService *auth.PasswordAuthService
	db              *gorm.DB
}

// NewProfileHandlers creates a new profile handlers instance
func NewProfileHandlers(db *gorm.DB) *ProfileHandlers {
	return &ProfileHandlers{
		userService:     auth.NewOAuthUserService(db),
		passwordService: auth.NewPasswordAuthService(db),
		db:              db,
	}
}

// UserProfileResponse represents user profile information
type UserProfileResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	RealmID     string     `json:"realm_id"`
	AvatarURL   string     `json:"avatar_url"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedTime time.Time  `json:"created_time"`
	Roles       []string   `json:"roles,omitempty"`
}

// RolesResponse represents user roles response
type RolesResponse struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	IsSuperAdmin bool    `json:"is_super_admin"`
	IsAdmin      bool    `json:"is_admin"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get the authenticated user's profile information
// @Tags profile
// @Produce json
// @Success 200 {object} UserProfileResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile [get]
func (h *ProfileHandlers) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Retrieve user from database
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile"})
		return
	}

	// Return profile information
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		RealmID:     user.RealmID,
		AvatarURL:   user.AvatarURL,
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedTime: user.CreatedTime,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile (email and name)
// @Tags profile
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update request"
// @Success 200 {object} UserProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile [put]
func (h *ProfileHandlers) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse request
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Retrieve current user
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Check if email is being changed and if new email already exists
	if req.Email != user.Email {
		existingUser, err := h.userService.GetUserByEmail(req.Email)
		if err != nil && err != auth.ErrUserNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email availability"})
			return
		}
		if existingUser != nil && existingUser.ID != user.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already in use by another user"})
			return
		}
	}

	// Update user profile
	user.Email = req.Email
	user.Name = req.Name
	user.UpdatedBy = user.Username
	user.UpdatedTime = time.Now()

	if err := h.db.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Return updated profile
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		RealmID:     user.RealmID,
		AvatarURL:   user.AvatarURL,
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedTime: user.CreatedTime,
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change the authenticated user's password
// @Tags profile
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile/change-password [post]
func (h *ProfileHandlers) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse request
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if user has a password (OAuth users don't have passwords)
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user.HashedPassword == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Cannot change password for OAuth users. Please use your OAuth provider to manage your password.",
		})
		return
	}

	// Change password using password service
	err = h.passwordService.ChangePassword(userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		if err == auth.ErrWeakPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == auth.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// GetRoles godoc
// @Summary Get current user roles
// @Description Get the authenticated user's roles (super_admin, admin, user)
// @Tags profile
// @Produce json
// @Success 200 {object} RolesResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile/roles [get]
func (h *ProfileHandlers) GetRoles(c *gin.Context) {
	// Get user ID and username from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	username, _ := c.Get("username")
	usernameStr := ""
	if username != nil {
		usernameStr = username.(string)
	}

	userIDStr := userID.(string)

	// Get user roles from database
	roles, err := authz.GetUserRoles(h.db, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
		return
	}

	// Check if user has specific roles
	isSuperAdmin := authz.IsSuperAdmin(h.db, userIDStr)
	isAdmin := authz.IsAdmin(h.db, userIDStr)

	// Return roles information
	c.JSON(http.StatusOK, RolesResponse{
		UserID:       userIDStr,
		Username:     usernameStr,
		Roles:        roles,
		IsSuperAdmin: isSuperAdmin,
		IsAdmin:      isAdmin,
	})
}
