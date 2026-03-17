package database

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB         *gorm.DB
	silentMode bool // Suppress log output to stdout/stderr (for MCP mode)
	dbLogger   *log.Logger
)

func init() {
	// Initialize database logger (default: stdout)
	dbLogger = log.New(os.Stdout, "", log.LstdFlags)
}

// SetSilentMode disables database logging to stdout/stderr (for MCP mode)
func SetSilentMode(silent bool) {
	silentMode = silent
	if silent {
		dbLogger.SetOutput(io.Discard)
	} else {
		dbLogger.SetOutput(os.Stdout)
	}
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, postgres, mysql
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Charset  string `mapstructure:"charset"`
	FilePath string `mapstructure:"file_path"` // for SQLite
}

// InitDB initializes database connection based on configuration
func InitDB() error {
	config := loadDatabaseConfig()

	var err error
	DB, err = connectDatabase(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate database schema
	if err := DB.AutoMigrate(models.GetAllModels()...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	// Initialize data
	if err := InitData(); err != nil {
		return fmt.Errorf("failed to initialize data: %w", err)
	}

	dbLogger.Printf("Successfully connected to %s database", config.Type)
	return nil
}

// loadDatabaseConfig loads database configuration from environment variables and config
func loadDatabaseConfig() *DatabaseConfig {
	config := &DatabaseConfig{
		Type:     getEnvOrDefault("DB_TYPE", "sqlite"),
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvIntOrDefault("DB_PORT", 5432),
		Username: getEnvOrDefault("DB_USERNAME", ""),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		Database: getEnvOrDefault("DB_NAME", "lazy_ai_coder"),
		SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		Charset:  getEnvOrDefault("DB_CHARSET", "utf8mb4"),
		FilePath: getEnvOrDefault("DB_FILE_PATH", "lazy_ai_coder.db"),
	}

	// Override with viper config if available
	if viper.IsSet("database.type") {
		config.Type = viper.GetString("database.type")
	}
	if viper.IsSet("database.host") {
		config.Host = viper.GetString("database.host")
	}
	if viper.IsSet("database.port") {
		config.Port = viper.GetInt("database.port")
	}
	if viper.IsSet("database.username") {
		config.Username = viper.GetString("database.username")
	}
	if viper.IsSet("database.password") {
		config.Password = viper.GetString("database.password")
	}
	if viper.IsSet("database.database") {
		config.Database = viper.GetString("database.database")
	}
	if viper.IsSet("database.ssl_mode") {
		config.SSLMode = viper.GetString("database.ssl_mode")
	}
	if viper.IsSet("database.charset") {
		config.Charset = viper.GetString("database.charset")
	}
	if viper.IsSet("database.file_path") {
		config.FilePath = viper.GetString("database.file_path")
	}

	return config
}

// connectDatabase establishes database connection based on type
func connectDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	switch config.Type {
	case "sqlite":
		return connectSQLite(config)
	case "postgres", "postgresql":
		return connectPostgreSQL(config)
	case "mysql":
		return connectMySQL(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// connectSQLite establishes SQLite connection
func connectSQLite(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := config.FilePath
	if dsn == "" {
		dsn = "lazy_ai_coder.db"
	}

	// Ensure the directory for the SQLite file exists so the file can be created
	if dir := filepath.Dir(dsn); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory %s: %w", dir, err)
		}
	}

	dbLogger.Printf("Connecting to SQLite database: %s", dsn)
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}

// connectPostgreSQL establishes PostgreSQL connection
func connectPostgreSQL(config *DatabaseConfig) (*gorm.DB, error) {
	if config.Port == 0 {
		config.Port = 5432
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	dbLogger.Printf("Connecting to PostgreSQL database: %s:%d/%s", config.Host, config.Port, config.Database)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// connectMySQL establishes MySQL connection
func connectMySQL(config *DatabaseConfig) (*gorm.DB, error) {
	if config.Port == 0 {
		config.Port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)

	dbLogger.Printf("Connecting to MySQL database: %s:%d/%s", config.Host, config.Port, config.Database)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// InitData initializes database with default data
func InitData() error {
	var count int64

	// Initialize LLM models
	DB.Model(&models.LLMModel{}).Count(&count)
	if count == 0 {
		defaultModels := []models.LLMModel{
			{ID: "llm_openai_gpt4", Name: "GPT-4", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "OpenAI GPT-4 - Most capable model for complex tasks", CreatedBy: "system"},
			{ID: "llm_openai_gpt4_turbo", Name: "GPT-4 Turbo", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4-turbo-preview", Temperature: 0.7, MaxTokens: 128000, IsEnabled: true, Description: "OpenAI GPT-4 Turbo - Faster with larger context window", CreatedBy: "system"},
			{ID: "llm_openai_gpt35", Name: "GPT-3.5 Turbo", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-3.5-turbo", Temperature: 0.7, MaxTokens: 16384, IsEnabled: true, Description: "OpenAI GPT-3.5 Turbo - Fast and cost-effective", CreatedBy: "system"},
			{ID: "llm_anthropic_claude3_opus", Name: "Claude 3 Opus", LLMType: "anthropic", BaseURL: "https://api.anthropic.com/v1", Model: "claude-3-opus-20240229", Temperature: 0.7, MaxTokens: 4096, IsEnabled: true, Description: "Anthropic Claude 3 Opus - Most intelligent model", CreatedBy: "system"},
			{ID: "llm_anthropic_claude3_sonnet", Name: "Claude 3.5 Sonnet", LLMType: "anthropic", BaseURL: "https://api.anthropic.com/v1", Model: "claude-3-5-sonnet-20241022", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "Anthropic Claude 3.5 Sonnet - Balanced performance", CreatedBy: "system"},
			{ID: "llm_google_gemini_pro", Name: "Gemini Pro", LLMType: "google", BaseURL: "https://generativelanguage.googleapis.com/v1beta", Model: "gemini-pro", Temperature: 0.7, MaxTokens: 32768, IsEnabled: true, Description: "Google Gemini Pro - Multimodal capabilities", CreatedBy: "system"},
			{ID: "llm_alibaba_qwen_max", Name: "Qwen Max", LLMType: "alibaba", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Model: "qwen-max", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "Alibaba Qwen Max - Excellent for Chinese and English", CreatedBy: "system"},
			{ID: "llm_deepseek_chat", Name: "DeepSeek Chat", LLMType: "deepseek", BaseURL: "https://api.deepseek.com/v1", Model: "deepseek-chat", Temperature: 0.7, MaxTokens: 32768, IsEnabled: true, Description: "DeepSeek Chat - Strong coding and reasoning", CreatedBy: "system"},
		}
		for _, m := range defaultModels {
			if err := DB.Create(&m).Error; err != nil {
				dbLogger.Printf("Warning: Failed to create default LLM model %s: %v", m.Name, err)
			}
		}
		dbLogger.Printf("Initialized database with %d default LLM models", len(defaultModels))
	}

	// Initialize prompts
	count = 0
	DB.Model(&models.Prompt{}).Count(&count)
	if count == 0 {
		// Load prompts from config
		var prompts []models.Prompt
		if err := viper.UnmarshalKey("prompts", &prompts); err != nil {
			return fmt.Errorf("unable to decode prompts into struct: %w", err)
		}

		if len(prompts) > 0 {
			result := DB.Create(&prompts)
			if result.Error != nil {
				return fmt.Errorf("failed to insert initial prompt data: %w", result.Error)
			}
			dbLogger.Printf("Initialized database with %d prompts", len(prompts))
		}
	}

	// Initialize system realm (for super admins)
	var systemRealm models.Realm
	err := DB.Where("id = ?", "system").First(&systemRealm).Error
	if err == gorm.ErrRecordNotFound {
		systemRealm = models.Realm{
			ID:          "system",
			Name:        "System",
			Description: "System realm for super administrators",
			CreatedBy:   "system",
		}
		if err := DB.Create(&systemRealm).Error; err != nil {
			return fmt.Errorf("failed to create system realm: %w", err)
		}
		dbLogger.Println("Created system realm")
	}

	// Initialize default roles (if not exist from migration)
	var roleCount int64
	DB.Model(&models.Role{}).Count(&roleCount)
	if roleCount == 0 {
		roles := []models.Role{
			{
				ID:          "role_super_admin",
				RealmID:     "system",
				Name:        "super_admin",
				Description: "Super administrator with full system access across all realms",
				CreatedBy:   "system",
			},
			{
				ID:          "role_admin",
				RealmID:     "system",
				Name:        "admin",
				Description: "Administrator with full access within their own realm",
				CreatedBy:   "system",
			},
			{
				ID:          "role_user",
				RealmID:     "system",
				Name:        "user",
				Description: "Regular user with standard permissions",
				CreatedBy:   "system",
			},
		}
		if err := DB.Create(&roles).Error; err != nil {
			return fmt.Errorf("failed to create default roles: %w", err)
		}
		dbLogger.Println("Created default roles")
	}

	// Initialize default super_admin user
	superAdminUsername := getEnvOrDefault("SUPER_ADMIN_USERNAME", "admin")
	superAdminPassword := getEnvOrDefault("SUPER_ADMIN_PASSWORD", "")
	superAdminEmail := getEnvOrDefault("SUPER_ADMIN_EMAIL", "admin@example.com")

	if superAdminPassword == "" {
		dbLogger.Println("Warning: SUPER_ADMIN_PASSWORD not set, skipping super_admin user creation")
		dbLogger.Println("To create super_admin user, set environment variable: SUPER_ADMIN_PASSWORD")
		return nil
	}

	// Check if super admin user already exists
	var existingUser models.User
	err = DB.Where("username = ?", superAdminUsername).First(&existingUser).Error
	if err == gorm.ErrRecordNotFound {
		// Create super admin user
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(superAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		userID := uuid.New().String()
		pwdHashStr := string(pwdHash)

		// Create user with personal realm
		personalRealmID := fmt.Sprintf("%s_realm", superAdminUsername)
		personalRealm := models.Realm{
			ID:          personalRealmID,
			Name:        fmt.Sprintf("%s's Realm", superAdminUsername),
			Description: fmt.Sprintf("Personal realm for super admin %s", superAdminUsername),
			CreatedBy:   "system",
		}
		if err := DB.Create(&personalRealm).Error; err != nil {
			return fmt.Errorf("failed to create personal realm: %w", err)
		}

		superAdminUser := models.User{
			ID:             userID,
			RealmID:        personalRealmID,
			Username:       superAdminUsername,
			HashedPassword: &pwdHashStr,
			Email:          superAdminEmail,
			IsActive:       true,
			CreatedBy:      "system",
		}

		if err := DB.Create(&superAdminUser).Error; err != nil {
			return fmt.Errorf("failed to create super_admin user: %w", err)
		}

		// Assign super_admin role to user
		userRole := models.UserRole{
			ID:        uuid.New().String(),
			UserID:    userID,
			RoleID:    "role_super_admin",
			CreatedBy: "system",
		}
		if err := DB.Create(&userRole).Error; err != nil {
			return fmt.Errorf("failed to assign super_admin role: %w", err)
		}

		dbLogger.Printf("Created super_admin user: %s with personal realm: %s", superAdminUsername, personalRealmID)
	} else if err == nil {
		// User exists, ensure they have super_admin role
		var existingRole models.UserRole
		err := DB.Where("user_id = ? AND role_id = ?", existingUser.ID, "role_super_admin").First(&existingRole).Error
		if err == gorm.ErrRecordNotFound {
			userRole := models.UserRole{
				ID:        uuid.New().String(),
				UserID:    existingUser.ID,
				RoleID:    "role_super_admin",
				CreatedBy: "system",
			}
			if err := DB.Create(&userRole).Error; err != nil {
				return fmt.Errorf("failed to assign super_admin role to existing user: %w", err)
			}
			dbLogger.Printf("Assigned super_admin role to existing user: %s", superAdminUsername)
		}
	}

	return nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
