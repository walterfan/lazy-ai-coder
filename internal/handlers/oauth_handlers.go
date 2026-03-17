package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/auth"
	"github.com/walterfan/lazy-ai-coder/internal/oauth"
)

// OAuthHandlers handles GitLab OAuth endpoints
type OAuthHandlers struct {
	gitlabOAuth *oauth.GitLabOAuthService
	userService *auth.OAuthUserService
	jwtService  *auth.SessionJWTService
	db          *gorm.DB
}

// NewOAuthHandlers creates a new OAuth handlers instance
func NewOAuthHandlers(db *gorm.DB) *OAuthHandlers {
	// Get OAuth configuration from environment
	gitlabBaseURL := os.Getenv("GITLAB_BASE_URL")
	if gitlabBaseURL == "" {
		gitlabBaseURL = "https://gitlab.com"
	}

	clientID := os.Getenv("GITLAB_OAUTH_APP_ID")
	clientSecret := os.Getenv("GITLAB_OAUTH_SECRET")
	redirectURI := os.Getenv("GITLAB_OAUTH_REDIRECT_URI")

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production" // Default for development
	}

	return &OAuthHandlers{
		gitlabOAuth: oauth.NewGitLabOAuthService(gitlabBaseURL, clientID, clientSecret, redirectURI),
		userService: auth.NewOAuthUserService(db),
		jwtService:  auth.NewSessionJWTService(jwtSecret, "ai-code-assistant", "ai-code-assistant"),
		db:          db,
	}
}

// generateState generates a random state parameter for CSRF protection
func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HandleLogin redirects to GitLab OAuth authorization page
func (h *OAuthHandlers) HandleLogin(c *gin.Context) {
	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Store state in session/cookie for validation
	// In production, store this in Redis or database with expiration
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	// Build GitLab OAuth URL
	authURL := h.gitlabOAuth.BuildAuthURL(state)

	// Redirect to GitLab
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback handles the OAuth callback from GitLab
func (h *OAuthHandlers) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Validate state to prevent CSRF
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	ctx := context.Background()

	// Exchange authorization code for access token
	tokenResp, err := h.gitlabOAuth.ExchangeCode(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange code: %v", err)})
		return
	}

	// Get user info from GitLab
	gitlabUser, err := h.gitlabOAuth.GetUser(ctx, tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user info: %v", err)})
		return
	}

	// Calculate token expiration time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Determine realm ID (for now, use "default" realm)
	// In production, you might want to map GitLab groups to realms
	realmID := "default"

	// Create or update user in database
	user, err := h.userService.CreateOrUpdateOAuthUser(
		gitlabUser.ID,
		gitlabUser.Username,
		gitlabUser.Email,
		gitlabUser.Name,
		gitlabUser.AvatarURL,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		expiresAt,
		realmID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create/update user: %v", err)})
		return
	}

	// Generate JWT session token (24 hours expiration)
	jwtToken, err := h.jwtService.GenerateToken(user.ID, user.RealmID, user.Username, user.Email, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate JWT: %v", err)})
		return
	}

	// Return JWT and user info to frontend
	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"realm_id":   user.RealmID,
		},
	})
}

// HandleGetCurrentUser returns the current authenticated user
func (h *OAuthHandlers) HandleGetCurrentUser(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"name":       user.Name,
		"avatar_url": user.AvatarURL,
		"realm_id":   user.RealmID,
	})
}

// HandleRefreshToken handles JWT token refresh
func (h *OAuthHandlers) HandleRefreshToken(c *gin.Context) {
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
	user, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return new token and user info
	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"realm_id":   user.RealmID,
		},
	})
}

// HandleLogout handles user logout
func (h *OAuthHandlers) HandleLogout(c *gin.Context) {
	// In a production system, you might want to:
	// 1. Blacklist the JWT token
	// 2. Revoke GitLab OAuth token
	// For now, just return success and let frontend handle token removal

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// OAuthMiddleware validates JWT tokens for OAuth-authenticated users
func (h *OAuthHandlers) OAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		// Validate JWT
		claims, err := h.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("realm_id", claims.RealmID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("auth_type", "oauth")

		c.Next()
	}
}

