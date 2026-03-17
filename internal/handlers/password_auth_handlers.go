package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/auth"
)

// PasswordAuthHandlers handles username/password authentication endpoints
type PasswordAuthHandlers struct {
	passwordService *auth.PasswordAuthService
	jwtService      *auth.SessionJWTService
	db              *gorm.DB
}

// NewPasswordAuthHandlers creates a new password auth handlers instance
func NewPasswordAuthHandlers(db *gorm.DB, jwtService *auth.SessionJWTService) *PasswordAuthHandlers {
	return &PasswordAuthHandlers{
		passwordService: auth.NewPasswordAuthService(db),
		jwtService:      jwtService,
		db:              db,
	}
}

// SignUpRequest represents the signup request body
type SignUpRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
}

// SignInRequest represents the signin request body
type SignInRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// UserProfile represents user information returned to client
type UserProfile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	RealmID  string `json:"realm_id"`
}

// HandleSignUp handles user registration with username/password
// @Summary Sign up with username/password
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Signup request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/signup [post]
func (h *PasswordAuthHandlers) HandleSignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// If name is not provided, use username
	if req.Name == "" {
		req.Name = req.Username
	}

	// Create user
	user, err := h.passwordService.SignUp(req.Username, req.Email, req.Password, req.Name)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, auth.ErrUsernameTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		if errors.Is(err, auth.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		if errors.Is(err, auth.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrInvalidUsername) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrInvalidEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token (24 hours expiration)
	token, err := h.jwtService.GenerateToken(user.ID, user.RealmID, user.Username, user.Email, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Return token and user info
	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User: UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			RealmID:  user.RealmID,
		},
	})
}

// HandleSignIn handles user login with username/password
// @Summary Sign in with username/password
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Signin request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/signin [post]
func (h *PasswordAuthHandlers) HandleSignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Authenticate user
	user, err := h.passwordService.SignIn(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Return generic error for other cases
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Generate JWT token (24 hours expiration)
	token, err := h.jwtService.GenerateToken(user.ID, user.RealmID, user.Username, user.Email, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Return token and user info
	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			RealmID:  user.RealmID,
		},
	})
}

// HandleRefreshToken handles JWT token refresh
// @Summary Refresh JWT token
// @Description Refresh an expired or expiring JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *PasswordAuthHandlers) HandleRefreshToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	// Extract Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	oldToken := authHeader[7:]

	// Refresh token with new 24-hour expiration
	newToken, err := h.jwtService.RefreshToken(oldToken, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
		return
	}

	// Validate to get user info
	claims, err := h.jwtService.ValidateToken(newToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token validation failed"})
		return
	}

	// Get user from database
	user, err := h.passwordService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return new token and user info
	c.JSON(http.StatusOK, AuthResponse{
		Token: newToken,
		User: UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			RealmID:  user.RealmID,
		},
	})
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// HandleChangePassword handles password change for authenticated users
// @Summary Change password
// @Description Change user's password (requires authentication)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/change-password [post]
func (h *PasswordAuthHandlers) HandleChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Change password
	err := h.passwordService.ChangePassword(userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		if errors.Is(err, auth.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
