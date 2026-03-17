package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	authService *AuthService
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// Login handles user authentication
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		case ErrUserInactive:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		case ErrInvalidRealm:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid realm"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(req)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		case ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		case ErrUserInactive:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Register handles user registration
func (h *AuthHandlers) Register(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get current user from context (for created_by field)
	currentUserID, exists := GetCurrentUser(c)
	if !exists {
		// For public registration, use a system user ID or nil
		currentUserID = uuid.Nil
	}

	user, err := h.authService.RegisterUser(req, currentUserID)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists in this realm"})
		case ErrInvalidRealm:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid realm"})
		case ErrPasswordTooShort:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password too short"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User registration failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandlers) GetProfile(c *gin.Context) {
	userID, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	realmID, exists := GetCurrentRealm(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := GetCurrentUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roles, exists := GetCurrentRoles(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	profile := gin.H{
		"user_id":  userID,
		"realm_id": realmID,
		"username": username,
		"email":    email,
		"roles":    roles,
	}

	c.JSON(http.StatusOK, profile)
}

// CheckPermission checks if the current user has permission for an action
func (h *AuthHandlers) CheckPermission(c *gin.Context) {
	userID, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	realmID, exists := GetCurrentRealm(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action parameter required"})
		return
	}

	resource := c.Query("resource")
	if resource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource parameter required"})
		return
	}

	// Build context from query parameters
	context := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "action" && key != "resource" && len(values) > 0 {
			context[key] = values[0]
		}
	}

	// Add user context
	context["user:id"] = userID.String()
	context["user:realm_id"] = realmID.String()
	context["user:username"] = c.GetString("username")
	context["user:email"] = c.GetString("email")
	context["user:roles"] = c.GetStringSlice("roles")

	allowed, err := h.authService.CheckPermission(userID, realmID, action, resource, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action":   action,
		"resource": resource,
		"allowed":  allowed,
	})
}

// Logout handles user logout (client-side token removal)
func (h *AuthHandlers) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// The server can't invalidate tokens, but we can return success
	// In production, you might want to implement a token blacklist
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// HealthCheck provides a health check endpoint
func (h *AuthHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth",
	})
}

// Admin handlers for user management
func (h *AuthHandlers) GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get users - not implemented yet"})
}

func (h *AuthHandlers) CreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create user - not implemented yet"})
}

func (h *AuthHandlers) GetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user - not implemented yet"})
}

func (h *AuthHandlers) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update user - not implemented yet"})
}

func (h *AuthHandlers) DeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete user - not implemented yet"})
}

// Admin handlers for role management
func (h *AuthHandlers) GetRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get roles - not implemented yet"})
}

func (h *AuthHandlers) CreateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create role - not implemented yet"})
}

func (h *AuthHandlers) GetRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get role - not implemented yet"})
}

func (h *AuthHandlers) UpdateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update role - not implemented yet"})
}

func (h *AuthHandlers) DeleteRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete role - not implemented yet"})
}

// Admin handlers for policy management
func (h *AuthHandlers) GetPolicies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get policies - not implemented yet"})
}

func (h *AuthHandlers) CreatePolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create policy - not implemented yet"})
}

func (h *AuthHandlers) GetPolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get policy - not implemented yet"})
}

func (h *AuthHandlers) UpdatePolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update policy - not implemented yet"})
}

func (h *AuthHandlers) DeletePolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete policy - not implemented yet"})
}

// Admin handlers for realm management
func (h *AuthHandlers) GetRealms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get realms - not implemented yet"})
}

func (h *AuthHandlers) CreateRealm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create realm - not implemented yet"})
}

func (h *AuthHandlers) GetRealm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get realm - not implemented yet"})
}

func (h *AuthHandlers) UpdateRealm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update realm - not implemented yet"})
}

func (h *AuthHandlers) DeleteRealm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete realm - not implemented yet"})
}
