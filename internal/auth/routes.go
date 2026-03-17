package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers authentication routes with the Gin router
func RegisterRoutes(router *gin.Engine, handlers *AuthHandlers, middleware *AuthMiddleware) {
	// Public routes (no authentication required)
	public := router.Group("/api/v1/auth")
	{
		public.POST("/login", handlers.Login)
		public.POST("/register", handlers.Register)
		public.POST("/refresh", handlers.RefreshToken)
		public.GET("/health", handlers.HealthCheck)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1/auth")
	protected.Use(middleware.Authenticate())
	{
		protected.GET("/profile", handlers.GetProfile)
		protected.GET("/permissions/check", handlers.CheckPermission)
		protected.POST("/logout", handlers.Logout)
	}

	// Admin routes (require admin role)
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.Authenticate())
	admin.Use(middleware.RequireRole("admin", "super_admin"))
	{
		// User management
		admin.GET("/users", handlers.GetUsers)          // List users
		admin.POST("/users", handlers.CreateUser)       // Create user
		admin.GET("/users/:id", handlers.GetUser)       // Get user
		admin.PUT("/users/:id", handlers.UpdateUser)    // Update user
		admin.DELETE("/users/:id", handlers.DeleteUser) // Delete user

		// Role management
		admin.GET("/roles", handlers.GetRoles)          // List roles
		admin.POST("/roles", handlers.CreateRole)       // Create role
		admin.GET("/roles/:id", handlers.GetRole)       // Get role
		admin.PUT("/roles/:id", handlers.UpdateRole)    // Update role
		admin.DELETE("/roles/:id", handlers.DeleteRole) // Delete role

		// Policy management
		admin.GET("/policies", handlers.GetPolicies)         // List policies
		admin.POST("/policies", handlers.CreatePolicy)       // Create policy
		admin.GET("/policies/:id", handlers.GetPolicy)       // Get policy
		admin.PUT("/policies/:id", handlers.UpdatePolicy)    // Update policy
		admin.DELETE("/policies/:id", handlers.DeletePolicy) // Delete policy

		// Realm management
		admin.GET("/realms", handlers.GetRealms)          // List realms
		admin.POST("/realms", handlers.CreateRealm)       // Create realm
		admin.GET("/realms/:id", handlers.GetRealm)       // Get realm
		admin.PUT("/realms/:id", handlers.UpdateRealm)    // Update realm
		admin.DELETE("/realms/:id", handlers.DeleteRealm) // Delete realm
	}
}

// RegisterProtectedRoutes registers routes that require authentication
// This can be used by other parts of the application to add protected routes
func RegisterProtectedRoutes(router *gin.RouterGroup, middleware *AuthMiddleware) {
	// Example of how to add protected routes with specific permissions
	router.Use(middleware.Authenticate())

	// Routes that require specific permissions
	projects := router.Group("/projects")
	{
		projects.GET("", middleware.RequirePermission("read:projects", "project:*"))
		projects.POST("", middleware.RequirePermission("create:project", "project:*"))
		projects.GET("/:id", middleware.RequirePermission("read:project", "project:*"))
		projects.PUT("/:id", middleware.RequirePermission("update:project", "project:*"))
		projects.DELETE("/:id", middleware.RequirePermission("delete:project", "project:*"))
	}

	prompts := router.Group("/prompts")
	{
		prompts.GET("", middleware.RequirePermission("read:prompts", "prompt:*"))
		prompts.POST("", middleware.RequirePermission("create:prompt", "prompt:*"))
		prompts.GET("/:id", middleware.RequirePermission("read:prompt", "prompt:*"))
		prompts.PUT("/:id", middleware.RequirePermission("update:prompt", "prompt:*"))
		prompts.DELETE("/:id", middleware.RequirePermission("delete:prompt", "prompt:*"))
	}

	// Routes that require specific roles
	adminOnly := router.Group("/admin")
	adminOnly.Use(middleware.RequireRole("admin", "super_admin"))
	{
		adminOnly.GET("/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin stats"})
		})
	}
}

// Example of how to use the middleware in your existing routes
func ExampleUsage(router *gin.Engine, middleware *AuthMiddleware) {
	// Public routes
	router.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Public endpoint"})
	})

	// Protected routes with authentication
	protected := router.Group("/protected")
	protected.Use(middleware.Authenticate())
	{
		protected.GET("/user-info", func(c *gin.Context) {
			userID, _ := GetCurrentUser(c)
			username, _ := GetCurrentUsername(c)
			c.JSON(200, gin.H{
				"user_id":  userID,
				"username": username,
			})
		})
	}

	// Routes with role-based access control
	admin := router.Group("/admin")
	admin.Use(middleware.Authenticate())
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin dashboard"})
		})
	}

	// Routes with permission-based access control
	projects := router.Group("/projects")
	projects.Use(middleware.Authenticate())
	{
		projects.GET("", middleware.RequirePermission("read:projects", "project:*"))
		projects.POST("", middleware.RequirePermission("create:project", "project:*"))
		projects.GET("/:id", middleware.RequirePermission("read:project", "project:*"))
		projects.PUT("/:id", middleware.RequirePermission("update:project", "project:*"))
		projects.DELETE("/:id", middleware.RequirePermission("delete:project", "project:*"))
	}
}
