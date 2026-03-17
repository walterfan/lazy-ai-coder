package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuth2Manager handles OAuth 2.0 flows
type OAuth2Manager struct {
	clients      map[string]*OAuthClient
	authCodes    map[string]*AuthCode
	accessTokens map[string]*AccessToken
}

// OAuthClient represents an OAuth 2.0 client application
type OAuthClient struct {
	ID           string    `json:"client_id"`
	Secret       string    `json:"client_secret"`
	Name         string    `json:"client_name"`
	RedirectURIs []string  `json:"redirect_uris"`
	Scopes       []string  `json:"scopes"`
	CreatedBy    uuid.UUID `json:"created_by"`
	CreatedTime  time.Time `json:"created_time"`
	IsActive     bool      `json:"is_active"`
}

// AuthCode represents an OAuth 2.0 authorization code
type AuthCode struct {
	Code        string    `json:"code"`
	ClientID    string    `json:"client_id"`
	UserID      uuid.UUID `json:"user_id"`
	RealmID     uuid.UUID `json:"realm_id"`
	RedirectURI string    `json:"redirect_uri"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expires_at"`
	Used        bool      `json:"used"`
}

// AccessToken represents an OAuth 2.0 access token
type AccessToken struct {
	Token     string    `json:"access_token"`
	TokenType string    `json:"token_type"`
	ExpiresIn int64     `json:"expires_in"`
	Scopes    []string  `json:"scopes"`
	UserID    uuid.UUID `json:"user_id"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}

// OAuth2Request represents an OAuth 2.0 authorization request
type OAuth2Request struct {
	ResponseType string `form:"response_type" binding:"required"`
	ClientID     string `form:"client_id" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	Scope        string `form:"scope"`
	State        string `form:"state"`
}

// OAuth2TokenRequest represents an OAuth 2.0 token request
type OAuth2TokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" binding:"required"`
	Scope        string `form:"scope"`
}

// NewOAuth2Manager creates a new OAuth 2.0 manager
func NewOAuth2Manager() *OAuth2Manager {
	return &OAuth2Manager{
		clients:      make(map[string]*OAuthClient),
		authCodes:    make(map[string]*AuthCode),
		accessTokens: make(map[string]*AccessToken),
	}
}

// RegisterClient registers a new OAuth 2.0 client
func (o *OAuth2Manager) RegisterClient(client *OAuthClient) error {
	if client.ID == "" {
		client.ID = generateRandomString(32)
	}
	if client.Secret == "" {
		client.Secret = generateRandomString(64)
	}

	client.CreatedTime = time.Now()
	o.clients[client.ID] = client
	return nil
}

// GetClient retrieves an OAuth client by ID
func (o *OAuth2Manager) GetClient(clientID string) (*OAuthClient, bool) {
	client, exists := o.clients[clientID]
	return client, exists
}

// ValidateClient validates client credentials
func (o *OAuth2Manager) ValidateClient(clientID, clientSecret string) bool {
	client, exists := o.clients[clientID]
	if !exists {
		return false
	}
	return client.Secret == clientSecret && client.IsActive
}

// CreateAuthCode creates a new authorization code
func (o *OAuth2Manager) CreateAuthCode(clientID string, userID, realmID uuid.UUID, redirectURI string, scopes []string) (*AuthCode, error) {
	code := &AuthCode{
		Code:        generateRandomString(32),
		ClientID:    clientID,
		UserID:      userID,
		RealmID:     realmID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Auth codes expire in 10 minutes
		Used:        false,
	}

	o.authCodes[code.Code] = code
	return code, nil
}

// ValidateAuthCode validates an authorization code
func (o *OAuth2Manager) ValidateAuthCode(code, clientID, redirectURI string) (*AuthCode, bool) {
	authCode, exists := o.authCodes[code]
	if !exists {
		return nil, false
	}

	if authCode.Used || time.Now().After(authCode.ExpiresAt) {
		return nil, false
	}

	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		return nil, false
	}

	return authCode, true
}

// CreateAccessToken creates a new access token
func (o *OAuth2Manager) CreateAccessToken(clientID string, userID uuid.UUID, scopes []string) *AccessToken {
	token := &AccessToken{
		Token:     generateRandomString(64),
		TokenType: "Bearer",
		ExpiresIn: 3600, // 1 hour
		Scopes:    scopes,
		UserID:    userID,
		ClientID:  clientID,
		CreatedAt: time.Now(),
	}

	o.accessTokens[token.Token] = token
	return token
}

// ValidateAccessToken validates an access token
func (o *OAuth2Manager) ValidateAccessToken(token string) (*AccessToken, bool) {
	accessToken, exists := o.accessTokens[token]
	if !exists {
		return nil, false
	}

	// Check if token has expired
	if time.Now().After(accessToken.CreatedAt.Add(time.Duration(accessToken.ExpiresIn) * time.Second)) {
		return nil, false
	}

	return accessToken, true
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// OAuth2Handlers provides HTTP handlers for OAuth 2.0 flows
type OAuth2Handlers struct {
	oauthManager *OAuth2Manager
	authService  *AuthService
}

// NewOAuth2Handlers creates new OAuth 2.0 handlers
func NewOAuth2Handlers(oauthManager *OAuth2Manager, authService *AuthService) *OAuth2Handlers {
	return &OAuth2Handlers{
		oauthManager: oauthManager,
		authService:  authService,
	}
}

// Authorize handles the OAuth 2.0 authorization endpoint
func (h *OAuth2Handlers) Authorize(c *gin.Context) {
	var req OAuth2Request
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Validate response type
	if req.ResponseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported response type"})
		return
	}

	// Validate client
	client, exists := h.oauthManager.GetClient(req.ClientID)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client"})
		return
	}

	// Validate redirect URI
	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == req.RedirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect URI"})
		return
	}

	// For now, redirect to login page
	// In a real implementation, you'd check if user is already authenticated
	c.JSON(http.StatusOK, gin.H{
		"message":      "Redirect to login page",
		"client_id":    req.ClientID,
		"redirect_uri": req.RedirectURI,
		"state":        req.State,
	})
}

// Token handles the OAuth 2.0 token endpoint
func (h *OAuth2Handlers) Token(c *gin.Context) {
	var req OAuth2TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Validate grant type
	if req.GrantType != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported grant type"})
		return
	}

	// Validate client credentials
	if !h.oauthManager.ValidateClient(req.ClientID, req.ClientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client credentials"})
		return
	}

	// Validate authorization code
	authCode, valid := h.oauthManager.ValidateAuthCode(req.Code, req.ClientID, req.RedirectURI)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization code"})
		return
	}

	// Mark code as used
	authCode.Used = true

	// Create access token
	accessToken := h.oauthManager.CreateAccessToken(req.ClientID, authCode.UserID, authCode.Scopes)

	c.JSON(http.StatusOK, accessToken)
}

// UserInfo handles the OAuth 2.0 user info endpoint
func (h *OAuth2Handlers) UserInfo(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenParts := []string{}
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenParts = []string{"Bearer", authHeader[7:]}
	}

	if len(tokenParts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
		return
	}

	tokenString := tokenParts[1]

	// Validate access token
	accessToken, valid := h.oauthManager.ValidateAccessToken(tokenString)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	// Get user information
	user, err := h.authService.userService.GetUserByID(accessToken.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":      user.ID.String(),
		"name":     user.Username,
		"email":    user.Email,
		"realm_id": user.RealmID.String(),
	})
}

// RegisterOAuthRoutes registers OAuth 2.0 routes
func RegisterOAuthRoutes(router *gin.Engine, handlers *OAuth2Handlers) {
	oauth := router.Group("/oauth2")
	{
		oauth.GET("/authorize", handlers.Authorize)
		oauth.POST("/token", handlers.Token)
		oauth.GET("/userinfo", handlers.UserInfo)
	}
}
