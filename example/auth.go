package example

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	ID           string    `json:"id" bson:"_id"`
	TenantID     string    `json:"tenant_id" bson:"tenant_id"`
	Username     string    `json:"username" bson:"username"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	Role         string    `json:"role" bson:"role"`     // admin, user, guest
	Status       string    `json:"status" bson:"status"` // active, inactive, suspended
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at" bson:"last_login_at"`
}

// Tenant 租户模型
type Tenant struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Domain      string    `json:"domain" bson:"domain"`
	Status      string    `json:"status" bson:"status"` // active, suspended, deleted
	Plan        string    `json:"plan" bson:"plan"`     // free, pro, enterprise
	MaxUsers    int       `json:"max_users" bson:"max_users"`
	MaxSessions int       `json:"max_sessions" bson:"max_sessions"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	jwtSecret []byte
	userDB    UserRepository
	tenantDB  TenantRepository
}

// UserRepository 用户数据访问接口
type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	GetUsersByTenant(tenantID string) ([]*User, error)
}

// TenantRepository 租户数据访问接口
type TenantRepository interface {
	GetByID(id string) (*Tenant, error)
	GetByDomain(domain string) (*Tenant, error)
	Create(tenant *Tenant) error
	Update(tenant *Tenant) error
}

// NewAuthService 创建认证服务
func NewAuthService(jwtSecret string, userDB UserRepository, tenantDB TenantRepository) *AuthService {
	return &AuthService{
		jwtSecret: []byte(jwtSecret),
		userDB:    userDB,
		tenantDB:  tenantDB,
	}
}

// RegisterUser 注册用户
func (as *AuthService) RegisterUser(tenantID, username, email, password string) (*User, error) {
	// 验证租户是否存在且活跃
	tenant, err := as.tenantDB.GetByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}
	if tenant.Status != "active" {
		return nil, fmt.Errorf("tenant is not active")
	}

	// 检查用户名是否已存在
	existingUser, _ := as.userDB.GetByUsername(username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, _ = as.userDB.GetByEmail(email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 检查租户用户数量限制
	users, err := as.userDB.GetUsersByTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users")
	}
	if len(users) >= tenant.MaxUsers {
		return nil, fmt.Errorf("tenant user limit exceeded")
	}

	// 加密密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	// 创建用户
	user := &User{
		ID:           generateID(),
		TenantID:     tenantID,
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "user",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = as.userDB.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	return user, nil
}

// Login 用户登录
func (as *AuthService) Login(email, password string) (string, error) {
	// 获取用户
	user, err := as.userDB.GetByEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", fmt.Errorf("user account is not active")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// 获取租户信息
	tenant, err := as.tenantDB.GetByID(user.TenantID)
	if err != nil {
		return "", fmt.Errorf("tenant not found")
	}
	if tenant.Status != "active" {
		return "", fmt.Errorf("tenant is not active")
	}

	// 更新最后登录时间
	user.LastLoginAt = time.Now()
	as.userDB.Update(user)

	// 生成JWT token
	token, err := as.generateToken(user, tenant)
	if err != nil {
		return "", fmt.Errorf("failed to generate token")
	}

	return token, nil
}

// ValidateToken 验证JWT token
func (as *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return as.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 验证用户状态
		user, err := as.userDB.GetByID(claims.UserID)
		if err != nil || user.Status != "active" {
			return nil, fmt.Errorf("user not active")
		}

		// 验证租户状态
		tenant, err := as.tenantDB.GetByID(claims.TenantID)
		if err != nil || tenant.Status != "active" {
			return nil, fmt.Errorf("tenant not active")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateToken 生成JWT token
func (as *AuthService) generateToken(user *User, tenant *Tenant) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.jwtSecret)
}

// generateID 生成唯一ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// AuthMiddleware 认证中间件
func (as *AuthService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}

		// 验证JWT token
		claims, err := as.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 获取用户和租户信息
		user, err := as.userDB.GetByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		tenant, err := as.tenantDB.GetByID(claims.TenantID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
			c.Abort()
			return
		}

		// 将信息存储到上下文中
		c.Set("user", user)
		c.Set("tenant", tenant)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func (as *AuthService) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*Claims)

		// 检查用户角色
		hasRole := false
		for _, role := range roles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenantAccess 租户访问验证中间件
func (as *AuthService) RequireTenantAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*Claims)
		user := c.MustGet("user").(*User)

		// 确保用户属于正确的租户
		if user.TenantID != claims.TenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "tenant access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
