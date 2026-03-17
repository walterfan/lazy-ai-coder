package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/walterfan/lazy-ai-coder/internal/assets"
	"github.com/walterfan/lazy-ai-coder/internal/auth"
	"github.com/walterfan/lazy-ai-coder/internal/chat"
	"github.com/walterfan/lazy-ai-coder/internal/chatrecord"
	"github.com/walterfan/lazy-ai-coder/internal/chatrecord/memory"
	"github.com/walterfan/lazy-ai-coder/internal/debug"
	"github.com/walterfan/lazy-ai-coder/internal/diagram"
	"github.com/walterfan/lazy-ai-coder/internal/handlers"
	"github.com/walterfan/lazy-ai-coder/internal/health"
	"github.com/walterfan/lazy-ai-coder/internal/mcp"
	"github.com/walterfan/lazy-ai-coder/internal/mem"
	"github.com/walterfan/lazy-ai-coder/internal/middleware"
	"github.com/walterfan/lazy-ai-coder/internal/smartprompt"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"gorm.io/gorm"
)

func getGitRepoConfig(c *gin.Context) {
	c.JSON(http.StatusOK, codeRepoConfig)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start a web server to handle LLM requests via HTTP",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration (required for web server)
		if err := InitConfig(); err != nil {
			panic(fmt.Sprintf("Failed to initialize config: %v", err))
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			panic(fmt.Sprintf("Failed to initialize database: %v", err))
		}
		defer database.CloseDB()

		r := setupRouter()

		// Start server on user-defined port
		addr := fmt.Sprintf(":%s", port)
		fmt.Printf("Starting web server on %s\n", addr)

		// Start periodic cleanup of expired sessions
		go func() {
			ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					mem.GetMemoryManager().CleanupExpiredSessions()
				}
			}
		}()

		if err := r.Run(addr); err != nil {
			panic(err)
		}
	},
}

// initChatRecordHandlers creates learning record (Coding Mate) handlers with optional LLM agent
func initChatRecordHandlers(db *gorm.DB) *handlers.ChatRecordHandlers {
	cfg := chatrecord.DefaultAgentConfig()
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("LLM_TEMPERATURE"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			cfg.Temperature = float32(f)
		}
	}
	if v := os.Getenv("LLM_MAX_TOKEN"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.MaxTokens = i
		}
	}

	if cfg.APIKey != "" {
		agent, err := chatrecord.NewEinoAgent(context.Background(), cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Code Mate agent init failed (submit will return error): %v\n", err)
		} else {
			sessionStore := memory.NewInMemorySessionStore(20, 30*time.Minute)
			return handlers.NewChatRecordHandlersWithAgentAndMemory(db, agent, sessionStore)
		}
	}
	return handlers.NewChatRecordHandlers(db)
}

// setupRouter configures and returns the Gin router with all routes and middleware
func setupRouter() *gin.Engine {
	r := gin.Default()
	db := database.GetDB()

	// Initialize services
	chatService := chat.NewChatService()
	diagramService := diagram.NewDiagramService()
	smartPromptService := smartprompt.NewSmartPromptService()

	// Initialize JWT service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}
	jwtService := auth.NewSessionJWTService(jwtSecret, "ai-code-assistant", "ai-code-assistant")

	// Initialize handlers
	oauthHandlers := handlers.NewOAuthHandlers(db)
	passwordAuthHandlers := handlers.NewPasswordAuthHandlers(db, jwtService)
	profileHandlers := handlers.NewProfileHandlers(db)
	chatHandlers := chat.NewChatHandlers(chatService)
	diagramHandlers := diagram.NewDiagramHandlers(diagramService)
	smartPromptHandlers := smartprompt.NewSmartPromptHandlers(smartPromptService)
	promptHandlers := handlers.NewPromptHandlers(db)
	projectHandlers := handlers.NewProjectHandlers(db)
	documentHandlers := handlers.NewDocumentHandlers(db)
	cursorRuleHandlers := handlers.NewCursorRuleHandlers(db)
	cursorCommandHandlers := handlers.NewCursorCommandHandlers(db)
	llmModelHandlers := handlers.NewLLMModelHandlers(db)
	ChatRecordHandlers := initChatRecordHandlers(db)
	userManagementHandlers := handlers.NewUserManagementHandlers(db)
	realmHandlers := handlers.NewRealmHandlers(db)
	debugHandlers := debug.NewDebugHandlers()
	healthChecker := health.NewHealthChecker("1.0.0") // TODO: Get version from build info

	// Assets (commands, rules, skills from assets folder)
	assetsRoot := os.Getenv("ASSETS_PATH")
	if assetsRoot == "" {
		assetsRoot = "assets"
	}
	assetsLoader, errAssets := assets.NewLoader(assetsRoot)
	if errAssets != nil {
		fmt.Fprintf(os.Stderr, "Assets loader init failed (assets endpoints will error): %v\n", errAssets)
		assetsLoader = nil
	}
	var assetsHandlers *handlers.AssetsHandlers
	if assetsLoader != nil {
		assetsHandlers = handlers.NewAssetsHandlers(assetsLoader)
	}

	// Initialize auth middleware
	userService := auth.NewOAuthUserService(db)
	flexibleAuth := handlers.NewFlexibleAuthMiddleware(db, jwtService, userService)

	// Apply middleware in order
	// 1. Request ID generation (must be first for correlation)
	r.Use(middleware.RequestIDMiddleware())

	// 2. Structured logging with request IDs
	r.Use(middleware.LoggingMiddleware())

	// 3. Metrics collection
	r.Use(middleware.MetricsMiddleware())

	// 4. Authentication
	// This supports BOTH authenticated users (OAuth/Password) and guest users
	r.Use(flexibleAuth.Middleware())

	// Public authentication endpoints (no auth required)
	// Username/Password authentication
	r.POST("/api/v1/auth/signup", passwordAuthHandlers.HandleSignUp)
	r.POST("/api/v1/auth/signin", passwordAuthHandlers.HandleSignIn)

	// OAuth authentication (GitLab)
	r.GET("/api/v1/auth/gitlab/login", oauthHandlers.HandleLogin)
	r.GET("/api/v1/auth/gitlab/callback", oauthHandlers.HandleCallback)

	// Token refresh endpoint (requires valid but potentially expiring token)
	r.POST("/api/v1/auth/refresh", passwordAuthHandlers.HandleRefreshToken)

	// Protected authentication endpoints (requires valid JWT token)
	authGroup := r.Group("/api/v1/auth")
	authGroup.Use(flexibleAuth.RequireAuth())
	{
		authGroup.GET("/user", oauthHandlers.HandleGetCurrentUser)
		authGroup.POST("/logout", oauthHandlers.HandleLogout)
		authGroup.POST("/change-password", passwordAuthHandlers.HandleChangePassword)
	}

	// Optional: Apply Basic Auth to admin-only routes if credentials provided
	// This is NOT applied globally - only to specific admin endpoints if needed
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username != "" && password != "" {
		// Example: Protect admin endpoints with Basic Auth
		// adminGroup := r.Group("/api/v1/admin")
		// adminGroup.Use(AuthMiddleware(username, password))
		// {
		//     // Admin routes here
		// }
		_ = username // Suppress unused variable warning for now
		_ = password
	}

	// Chat endpoints
	r.GET("/api/v1/stream", chatHandlers.HandleWebSocket)
	r.POST("/api/v1/process", chatHandlers.HandleChatRequest)

	// Diagram endpoints
	r.POST("/api/v1/draw", diagramHandlers.HandleDrawRequest)

	// Smart Prompt Generator endpoints
	r.POST("/api/v1/smart-prompt/generate", smartPromptHandlers.HandleSmartPromptGenerate)
	r.GET("/api/v1/smart-prompt/presets", smartPromptHandlers.HandleGetPresets)

	// Framework endpoints
	r.GET("/api/v1/smart-prompt/frameworks", smartPromptHandlers.HandleGetFrameworks)
	r.GET("/api/v1/smart-prompt/frameworks/:id", smartPromptHandlers.HandleGetFramework)
	r.POST("/api/v1/smart-prompt/frameworks", smartPromptHandlers.HandleCreateFramework)
	r.PUT("/api/v1/smart-prompt/frameworks/:id", smartPromptHandlers.HandleUpdateFramework)
	r.DELETE("/api/v1/smart-prompt/frameworks/:id", smartPromptHandlers.HandleDeleteFramework)

	// Template endpoints
	r.GET("/api/v1/smart-prompt/templates/categories", smartPromptHandlers.HandleGetTemplateCategories)
	r.GET("/api/v1/smart-prompt/templates", smartPromptHandlers.HandleGetTemplates)
	r.GET("/api/v1/smart-prompt/templates/:id", smartPromptHandlers.HandleGetTemplate)
	r.POST("/api/v1/smart-prompt/templates/:id/use", smartPromptHandlers.HandleUseTemplate)
	r.POST("/api/v1/smart-prompt/templates", smartPromptHandlers.HandleCreateTemplate)
	r.PUT("/api/v1/smart-prompt/templates/:id", smartPromptHandlers.HandleUpdateTemplate)
	r.DELETE("/api/v1/smart-prompt/templates/:id", smartPromptHandlers.HandleDeleteTemplate)

	// Refinement endpoints
	r.POST("/api/v1/smart-prompt/refine", smartPromptHandlers.HandleRefinePrompt)
	r.POST("/api/v1/smart-prompt/quick-refine", smartPromptHandlers.HandleQuickRefine)
	r.POST("/api/v1/smart-prompt/refine-with-requirements", smartPromptHandlers.HandleRefineWithRequirements)

	// Generation from framework
	r.POST("/api/v1/smart-prompt/generate-from-framework", smartPromptHandlers.HandleGenerateFromFramework)

	// Auto-fill framework fields with LLM
	r.POST("/api/v1/smart-prompt/auto-fill-fields", smartPromptHandlers.HandleAutoFillFields)

	// Public endpoints - Read operations (Guest + Authenticated)
	r.GET("/api/v1/prompts", promptHandlers.ListPrompts)
	r.GET("/api/v1/prompts/:id", promptHandlers.GetPrompt)
	r.GET("/api/v1/cursor-rules", cursorRuleHandlers.ListCursorRules)
	r.GET("/api/v1/cursor-rules/:id", cursorRuleHandlers.GetCursorRule)
	r.GET("/api/v1/cursor-rules/:id/export", cursorRuleHandlers.ExportCursorRule)
	r.GET("/api/v1/cursor-commands", cursorCommandHandlers.ListCursorCommands)
	r.GET("/api/v1/cursor-commands/:id", cursorCommandHandlers.GetCursorCommand)
	r.GET("/api/v1/cursor-commands/:id/export", cursorCommandHandlers.ExportCursorCommand)
	r.GET("/api/v1/projects", projectHandlers.ListProjects)
	r.GET("/api/v1/projects/export", projectHandlers.ExportProjects)
	r.GET("/api/v1/projects/:id", projectHandlers.GetProject)

	// Document endpoints - Read operations (Guest + Authenticated)
	r.GET("/api/v1/documents", documentHandlers.ListDocuments)
	r.GET("/api/v1/documents/stats", documentHandlers.GetDocumentStats)
	r.GET("/api/v1/documents/chunks", documentHandlers.GetDocumentChunks)
	r.GET("/api/v1/documents/:id", documentHandlers.GetDocument)

	// LLM Model endpoints - Read operations (Guest + Authenticated)
	r.GET("/api/v1/llm-models", llmModelHandlers.ListLLMModels)
	r.GET("/api/v1/llm-models/default", llmModelHandlers.GetDefaultLLMModel)
	r.GET("/api/v1/llm-models/:id", llmModelHandlers.GetLLMModel)

	// Assets (commands, rules, skills - search and download)
	if assetsHandlers != nil {
		r.GET("/api/v1/assets", assetsHandlers.ListAssets)
		r.GET("/api/v1/assets/download", assetsHandlers.DownloadAsset)
		r.GET("/api/v1/assets/download-skill", assetsHandlers.DownloadSkillZip)
	}

	// Protected endpoints - Create/Update/Delete operations (Authenticated only)
	protected := r.Group("/api/v1")
	protected.Use(handlers.RequireAuthenticated())
	{
		// Profile management
		protected.GET("/profile", profileHandlers.GetProfile)
		protected.PUT("/profile", profileHandlers.UpdateProfile)
		protected.POST("/profile/change-password", profileHandlers.ChangePassword)
		protected.GET("/profile/roles", profileHandlers.GetRoles)

		// Prompts CUD
		protected.POST("/prompts", promptHandlers.CreatePrompt)
		protected.PUT("/prompts/:id", promptHandlers.UpdatePrompt)
		protected.DELETE("/prompts/:id", promptHandlers.DeletePrompt)

		// Cursor Rules CUD
		protected.POST("/cursor-rules", cursorRuleHandlers.CreateCursorRule)
		protected.PUT("/cursor-rules/:id", cursorRuleHandlers.UpdateCursorRule)
		protected.DELETE("/cursor-rules/:id", cursorRuleHandlers.DeleteCursorRule)
		protected.POST("/cursor-rules/generate", cursorRuleHandlers.GenerateCursorRule)
		protected.POST("/cursor-rules/:id/refine", cursorRuleHandlers.RefineCursorRule)
		protected.POST("/cursor-rules/import", cursorRuleHandlers.ImportCursorRule)
		protected.POST("/cursor-rules/validate", cursorRuleHandlers.ValidateCursorRule)

		// Cursor Commands CUD
		protected.POST("/cursor-commands", cursorCommandHandlers.CreateCursorCommand)
		protected.PUT("/cursor-commands/:id", cursorCommandHandlers.UpdateCursorCommand)
		protected.DELETE("/cursor-commands/:id", cursorCommandHandlers.DeleteCursorCommand)
		protected.POST("/cursor-commands/generate", cursorCommandHandlers.GenerateCursorCommand)
		protected.POST("/cursor-commands/:id/refine", cursorCommandHandlers.RefineCursorCommand)
		protected.POST("/cursor-commands/import", cursorCommandHandlers.ImportCursorCommand)

		// Projects CUD
		protected.POST("/projects", projectHandlers.CreateProject)
		protected.PUT("/projects/:id", projectHandlers.UpdateProject)
		protected.DELETE("/projects/:id", projectHandlers.DeleteProject)
		protected.POST("/projects/import", projectHandlers.ImportProjects)

		// Documents CUD (URL loading, file upload, text input, delete)
		protected.POST("/documents/load-url", documentHandlers.LoadFromURL)
		protected.POST("/documents/upload", documentHandlers.UploadFiles)
		protected.POST("/documents/create-from-text", documentHandlers.CreateFromText)
		protected.DELETE("/documents/:id", documentHandlers.DeleteDocument)
		protected.POST("/documents/delete-by-path", documentHandlers.DeleteDocumentByPath)

		// LLM Models CUD
		protected.POST("/llm-models", llmModelHandlers.CreateLLMModel)
		protected.PUT("/llm-models/:id", llmModelHandlers.UpdateLLMModel)
		protected.DELETE("/llm-models/:id", llmModelHandlers.DeleteLLMModel)
		protected.POST("/llm-models/:id/default", llmModelHandlers.SetDefaultLLMModel)
		protected.POST("/llm-models/:id/toggle", llmModelHandlers.ToggleLLMModelEnabled)

		// Learning Record endpoints
		protected.POST("/chat-record/submit", ChatRecordHandlers.HandleSubmit)
		protected.POST("/chat-record/confirm", ChatRecordHandlers.HandleConfirm)
		protected.GET("/chat-record/list", ChatRecordHandlers.HandleList)
		protected.GET("/chat-record/stats", ChatRecordHandlers.HandleStats)
		protected.GET("/chat-record/:id", ChatRecordHandlers.HandleGet)
		protected.DELETE("/chat-record/:id", ChatRecordHandlers.HandleDelete)
	}

	// Admin-only endpoints (super_admin role required)
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(handlers.LoadUserRoles(db))
	adminGroup.Use(handlers.RequireSuperAdmin(db))
	{
		// User management
		adminGroup.GET("/pending-users", userManagementHandlers.GetPendingUsers)
		adminGroup.POST("/users/:id/approve", userManagementHandlers.ApproveUser)
		adminGroup.POST("/users/:id/reject", userManagementHandlers.RejectUser)
		adminGroup.GET("/users", userManagementHandlers.GetAllUsers)
		adminGroup.PUT("/users/:id/realm", userManagementHandlers.UpdateUserRealm)
		adminGroup.PUT("/users/:id/role", userManagementHandlers.UpdateUserRole)
		adminGroup.POST("/users/:id/deactivate", userManagementHandlers.DeactivateUser)

		// Realm management
		adminGroup.GET("/realms", realmHandlers.GetAllRealms)
		adminGroup.GET("/realms/:id", realmHandlers.GetRealmByID)
		adminGroup.POST("/realms", realmHandlers.CreateRealm)
		adminGroup.PUT("/realms/:id", realmHandlers.UpdateRealm)
		adminGroup.DELETE("/realms/:id", realmHandlers.DeleteRealm)
		adminGroup.GET("/realms/:id/users", realmHandlers.GetUsersInRealm)
	}

	// GitLab config endpoint
	r.GET("/api/v1/gitlab_config", getGitRepoConfig)

	// Memory management endpoints
	r.GET("/api/v1/sessions/:sessionId", func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(sessionID)

		c.JSON(http.StatusOK, gin.H{
			"session":  session.GetStats(),
			"messages": session.GetMessages(),
		})
	})

	r.DELETE("/api/v1/sessions/:sessionId", func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		memoryManager := mem.GetMemoryManager()
		memoryManager.DeleteSession(sessionID)

		c.JSON(http.StatusOK, gin.H{"message": "Session cleared"})
	})

	r.GET("/api/v1/sessions", func(c *gin.Context) {
		memoryManager := mem.GetMemoryManager()
		sessions := memoryManager.GetAllSessions()

		c.JSON(http.StatusOK, gin.H{"sessions": sessions})
	})

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, healthChecker.BasicHealthCheck())
	})
	r.GET("/health/detailed", func(c *gin.Context) {
		c.JSON(http.StatusOK, healthChecker.DetailedHealthCheck())
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Debug endpoints (memory management)
	debugGroup := r.Group("/api/v1/debug")
	{
		debugGroup.GET("/memory/sessions", debugHandlers.ListSessions)
		debugGroup.GET("/memory/sessions/:session_id", debugHandlers.GetSession)
		debugGroup.DELETE("/memory/sessions/:session_id", debugHandlers.DeleteSession)
		debugGroup.POST("/memory/sessions/:session_id/summarize", debugHandlers.TriggerSummarization)
		debugGroup.GET("/memory/stats", debugHandlers.GetMemoryStats)
		debugGroup.POST("/memory/cleanup", debugHandlers.CleanupExpiredSessions)
	}

	// MCP server integration (HTTP-based)
	mcpServer := mcp.NewHTTPServer()
	// Initialize database and load prompts/resources
	if err := mcpServer.InitializeWithDB(); err != nil {
		fmt.Printf("Warning: Failed to initialize MCP prompts/resources: %v\n", err)
		// Continue anyway - tools will still work
	}
	mcpServer.RegisterAllTools()
	mcpServer.SetupRoutes(r)

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files from web/dist for production Vue app
	r.Static("/assets", "./web/dist/assets")
	r.Static("/images", "./web/images")

	// SPA: serve index.html for the root and all Vue Router paths
	serveIndex := func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./web/dist/index.html")
	}
	r.GET("/", serveIndex)

	// Catch-all for client-side routes so direct navigation / bookmark / refresh works
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/mcp/") || strings.HasPrefix(path, "/swagger/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}

		if _, err := os.Stat("./web/dist/index.html"); os.IsNotExist(err) {
			c.String(http.StatusNotFound, "Frontend not found. Please build the web application (cd web && npm run build).")
			return
		}

		serveIndex(c)
	})

	return r
}

func AuthMiddleware(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != username || pass != password {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
func init() {
	webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Specify custom port for the web server")
	rootCmd.AddCommand(webCmd)
}
