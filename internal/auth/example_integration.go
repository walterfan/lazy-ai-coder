package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExampleIntegration shows how to integrate the auth system with your existing app
func ExampleIntegration() {
	// This is an example of how to integrate the auth system
	// with your existing lazy-ai-coder application

	_ = gin.Default() // router variable for example purposes

	// Initialize auth components (you'll need to implement the interfaces)
	// userService := &YourUserService{}
	// policyService := &YourPolicyService{}

	// passwordManager := NewPasswordManager(12)
	// jwtManager, _ := NewJWTManager("private.pem", "public.pem", "lazy-ai-coder", "web-client")
	// permissionEngine := NewPermissionEngine(policyService)
	// authService := NewAuthService(userService, passwordManager, jwtManager, permissionEngine)

	// handlers := NewAuthHandlers(authService)
	// middleware := NewAuthMiddleware(authService)

	// Register auth routes
	// RegisterRoutes(router, handlers, middleware)

	// Example: Protect your existing LLM endpoints
	// llmGroup := router.Group("/api/v1/llm")
	// llmGroup.Use(middleware.Authenticate()) // Require authentication
	// {
	//     llmGroup.POST("/chat", middleware.RequirePermission("chat:send", "llm:*"), chatHandler)
	//     llmGroup.GET("/history", middleware.RequirePermission("chat:read", "llm:*"), historyHandler)
	// }

	// Example: Protect your existing command endpoints
	// commandsGroup := router.Group("/api/v1/commands")
	// commandsGroup.Use(middleware.Authenticate())
	// {
	//     commandsGroup.GET("", middleware.RequirePermission("commands:read", "command:*"), listCommandsHandler)
	//     commandsGroup.POST("", middleware.RequirePermission("commands:create", "command:*"), createCommandHandler)
	// }

	// Example: Protect your existing project endpoints
	// projectsGroup := router.Group("/api/v1/projects")
	// projectsGroup.Use(middleware.Authenticate())
	// {
	//     projectsGroup.GET("", middleware.RequirePermission("projects:read", "project:*"), listProjectsHandler)
	//     projectsGroup.GET("/:id", middleware.RequirePermission("projects:read", "project:*"), getProjectHandler)
	// }

	// Example: Admin-only endpoints
	// adminGroup := router.Group("/api/v1/admin")
	// adminGroup.Use(middleware.Authenticate())
	// adminGroup.Use(middleware.RequireRole("admin", "super_admin"))
	// {
	//     adminGroup.GET("/users", listUsersHandler)
	//     adminGroup.POST("/users", createUserHandler)
	//     adminGroup.GET("/stats", statsHandler)
	// }

	// Example: Realm-specific endpoints
	// realmGroup := router.Group("/api/v1/realm/:realm_id")
	// realmGroup.Use(middleware.Authenticate())
	// realmGroup.Use(middleware.RequireRealm(getRealmIDFromParam)) // Custom middleware
	// {
	//     realmGroup.GET("/settings", getRealmSettingsHandler)
	//     realmGroup.PUT("/settings", updateRealmSettingsHandler)
	// }

	// Example: Conditional permissions based on resource ownership
	// projectGroup := router.Group("/api/v1/projects/:project_id")
	// projectGroup.Use(middleware.Authenticate())
	// {
	//     // Read access for all authenticated users in the realm
	//     projectGroup.GET("", middleware.RequirePermission("projects:read", "project:*"), getProjectHandler)
	//
	//     // Write access only for project owners or admins
	//     projectGroup.PUT("", middleware.RequirePermission("projects:update", "project:*"), updateProjectHandler)
	//     projectGroup.DELETE("", middleware.RequirePermission("projects:delete", "project:*"), deleteProjectHandler)
	// }

	// Example: WebSocket authentication
	// router.GET("/ws", middleware.OptionalAuth(), websocketHandler)

	// Example: Public endpoints with optional authentication
	// router.GET("/public/info", middleware.OptionalAuth(), publicInfoHandler)

	// Example: Rate limiting for auth endpoints
	// authGroup := router.Group("/api/v1/auth")
	// authGroup.Use(rateLimitMiddleware) // Your rate limiting middleware
	// {
	//     authGroup.POST("/login", handlers.Login)
	//     authGroup.POST("/register", handlers.Register)
	// }

	// Example: CORS and security headers
	// router.Use(corsMiddleware)
	// router.Use(securityHeadersMiddleware)

	// Example: Health check endpoint (no auth required)
	// router.GET("/health", healthCheckHandler)

	// Example: Metrics endpoint (admin only)
	// router.GET("/metrics", middleware.Authenticate(), middleware.RequireRole("admin"), metricsHandler)

	// Example: API documentation (public)
	// router.GET("/docs", serveDocsHandler)

	// Example: Static files (public)
	// router.Static("/static", "./static")

	// Example: Error handling middleware
	// router.Use(errorHandlingMiddleware)

	// Example: Request logging middleware
	// router.Use(requestLoggingMiddleware)

	// Example: Graceful shutdown
	// go func() {
	//     if err := router.Run(":8080"); err != nil {
	//         log.Fatal("Failed to start server:", err)
	//     }
	// }()

	// // Wait for shutdown signal
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	//
	// log.Println("Shutting down server...")
	// // Graceful shutdown logic here
}

// Example of how to implement the UserService interface
type ExampleUserService struct {
	// Your database connection, cache, etc.
}

func (s *ExampleUserService) GetUserByID(userID uuid.UUID) (*User, error) {
	// Implement database query to get user by ID
	// Return auth.User struct
	return nil, nil
}

func (s *ExampleUserService) GetUserByUsername(username string, realmID uuid.UUID) (*User, error) {
	// Implement database query to get user by username and realm
	// Return auth.User struct
	return nil, nil
}

func (s *ExampleUserService) GetUserRoles(userID uuid.UUID) ([]*Role, error) {
	// Implement database query to get user roles
	// Return slice of auth.Role structs
	return nil, nil
}

func (s *ExampleUserService) CreateUser(user *User) error {
	// Implement database insert for new user
	return nil
}

func (s *ExampleUserService) UpdateUser(user *User) error {
	// Implement database update for existing user
	return nil
}

func (s *ExampleUserService) DeleteUser(userID uuid.UUID) error {
	// Implement database delete for user
	return nil
}

// Example of how to implement the PolicyService interface
type ExamplePolicyService struct {
	// Your database connection, cache, etc.
}

func (s *ExamplePolicyService) GetUserPolicies(userID, realmID uuid.UUID) ([]*Policy, error) {
	// Implement database query to get all policies for a user
	// This should include:
	// 1. Direct user policies
	// 2. Role-based policies
	// 3. Resource-based policies
	// Return slice of auth.Policy structs
	return nil, nil
}

func (s *ExamplePolicyService) GetPolicyStatements(policyID uuid.UUID) ([]*Statement, error) {
	// Implement database query to get statements for a policy
	// Return slice of auth.Statement structs
	return nil, nil
}

// Example middleware for realm validation
func getRealmIDFromParam(c *gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		realmIDStr := c.Param("realm_id")
		realmID, err := uuid.Parse(realmIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid realm ID"})
			c.Abort()
			return
		}

		// Set realm ID in context for RequireRealm middleware
		c.Set("requested_realm_id", realmID)
		c.Next()
	}
}

// Example of how to use the auth system in your existing handlers
func ExampleProtectedHandler(c *gin.Context) {
	// Get current user information
	userID, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	realmID, exists := GetCurrentRealm(c)
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := GetCurrentUsername(c)
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	// Use the authenticated user information
	c.JSON(200, gin.H{
		"message":  "Hello, " + username,
		"user_id":  userID,
		"realm_id": realmID,
	})
}

// Example of how to check permissions programmatically
func ExamplePermissionCheck(c *gin.Context) {
	userID, _ := GetCurrentUser(c)
	realmID, _ := GetCurrentRealm(c)

	// Check if user can read a specific project
	projectID := c.Param("project_id")
	_ = map[string]interface{}{ // context variable for example purposes
		"project:id":    projectID,
		"user:id":       userID.String(),
		"user:realm_id": realmID.String(),
	}

	// You would inject the authService here
	// allowed, err := authService.CheckPermission(userID, realmID, "read:project", "project:"+projectID, context)
	// if err != nil {
	//     c.JSON(500, gin.H{"error": "Failed to check permissions"})
	//     return
	// }
	//
	// if !allowed {
	//     c.JSON(403, gin.H{"error": "Access denied"})
	//     return
	// }

	c.JSON(200, gin.H{"message": "Permission check passed"})
}
