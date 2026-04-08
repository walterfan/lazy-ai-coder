package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/auth"
	"github.com/walterfan/lazy-ai-coder/internal/services"
)

// FlexibleAuthMiddleware supports BOTH authenticated users (OAuth/Password) and guest users
// - Authenticated users: Have JWT token, full CRUD access, realm-scoped
// - Guest users: No token, read-only access, use "guest" realm
type FlexibleAuthMiddleware struct {
	jwtService  *auth.SessionJWTService
	userService *auth.OAuthUserService
	db          *gorm.DB
}

// NewFlexibleAuthMiddleware creates a new flexible auth middleware
func NewFlexibleAuthMiddleware(db *gorm.DB, jwtService *auth.SessionJWTService, userService *auth.OAuthUserService) *FlexibleAuthMiddleware {
	return &FlexibleAuthMiddleware{
		jwtService:  jwtService,
		userService: userService,
		db:          db,
	}
}

// Middleware returns the Gin middleware handler
// Sets authentication context for all requests
func (m *FlexibleAuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to authenticate with JWT token (OAuth or password)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := authHeader[7:]

		// Validate JWT
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[AUTH] JWT validation failed for %s: %v\n", c.Request.URL.Path, err)
		} else {
			user, err := m.userService.GetUserByID(claims.UserID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[AUTH] User lookup failed for userID=%s path=%s: %v\n", claims.UserID, c.Request.URL.Path, err)
			} else {
				c.Set("is_authenticated", true)
				c.Set("auth_type", "oauth")
				c.Set("user_id", user.ID)
				c.Set("realm_id", user.RealmID)
				c.Set("username", user.Username)
				c.Set("email", user.Email)
				if user.GitLabAccessToken != nil {
					c.Set("gitlab_token", *user.GitLabAccessToken)
				}
				c.Next()
				return
			}
		}
		}

		// No valid JWT token - guest mode (read-only)
		// Guest users use credentials from request body (Settings.GITLAB_TOKEN, LLM_API_KEY)
		c.Set("is_authenticated", false)
		c.Set("auth_type", "guest")
		c.Set("user_id", "")           // No user ID for guests
		c.Set("realm_id", "guest")     // Guest realm
		c.Set("username", "guest")
		c.Set("email", "")
		// gitlab_token will be extracted from request body by handlers
		c.Next()
	}
}

// RequireAuth is a stricter middleware that requires authentication
// Use this for endpoints that must have a valid user
func (m *FlexibleAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try OAuth JWT first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := authHeader[7:]

			claims, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				c.Abort()
				return
			}

			// Get user from database
			user, err := m.userService.GetUserByID(claims.UserID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}

			// Set context
			c.Set("auth_type", "oauth")
			c.Set("user_id", user.ID)
			c.Set("realm_id", user.RealmID)
			c.Set("username", user.Username)
			c.Set("email", user.Email)
			if user.GitLabAccessToken != nil {
				c.Set("gitlab_token", *user.GitLabAccessToken)
			}

			c.Next()
			return
		}

		// No valid auth found
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
	}
}

// GetUserContext extracts user context from Gin context
func GetUserContext(c *gin.Context) (authType, userID, realmID, username, email string) {
	authTypeVal, _ := c.Get("auth_type")
	userIDVal, _ := c.Get("user_id")
	realmIDVal, _ := c.Get("realm_id")
	usernameVal, _ := c.Get("username")
	emailVal, _ := c.Get("email")

	authType, _ = authTypeVal.(string)
	userID, _ = userIDVal.(string)
	realmID, _ = realmIDVal.(string)
	username, _ = usernameVal.(string)
	email, _ = emailVal.(string)

	return authType, userID, realmID, username, email
}

// GetGitLabToken extracts GitLab token from context or request body
func GetGitLabToken(c *gin.Context) string {
	// Try to get from context (authenticated user's stored token)
	if token, exists := c.Get("gitlab_token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}

	// For guest mode, token should be in request body
	// This will be handled by individual handlers that parse the request
	return ""
}

// RequireAuthenticated is a middleware that blocks guest users
// Use this to protect Create/Update/Delete endpoints
func RequireAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This action requires authentication. Guest mode is read-only. Please sign up or sign in to create, update, or delete resources.",
				"hint":  "Visit /settings to sign up or sign in",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSuperAdmin is a middleware that requires super_admin role
// Use this to protect super admin-only endpoints (user approval, realm management, etc.)
func RequireSuperAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check authentication first
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Check if user is super admin
		roleService := services.NewRoleService(db)
		isSuperAdmin, err := roleService.IsSuperAdmin(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check user role",
			})
			c.Abort()
			return
		}

		if !isSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Super admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a middleware that requires admin or super_admin role
// Use this to protect admin endpoints
func RequireAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check authentication first
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Check if user is admin or super admin
		roleService := services.NewRoleService(db)
		isAdmin, err := roleService.IsAdmin(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check user role",
			})
			c.Abort()
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoadUserRoles is a middleware that loads user roles into context
// Should be used after FlexibleAuthMiddleware for authenticated users
func LoadUserRoles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only load roles for authenticated users
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			c.Next()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Load user roles
		roleService := services.NewRoleService(db)
		roleNames, err := roleService.GetUserRoleNames(userID.(string))
		if err == nil && len(roleNames) > 0 {
			c.Set("roles", roleNames)

			// Set convenience flags
			isSuperAdmin, _ := roleService.IsSuperAdmin(userID.(string))
			isAdmin, _ := roleService.IsAdmin(userID.(string))
			c.Set("is_super_admin", isSuperAdmin)
			c.Set("is_admin", isAdmin)
		}

		c.Next()
	}
}

