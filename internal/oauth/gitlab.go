package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GitLabUser represents a GitLab user from the OAuth API
type GitLabUser struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// GitLabTokenResponse represents the response from GitLab token endpoint
type GitLabTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int64  `json:"created_at"`
	Scope        string `json:"scope"`
}

// GitLabOAuthService handles GitLab OAuth 2.0 flow
type GitLabOAuthService struct {
	GitLabBaseURL string
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	httpClient    *http.Client
}

// NewGitLabOAuthService creates a new GitLab OAuth service
func NewGitLabOAuthService(gitlabBaseURL, clientID, clientSecret, redirectURI string) *GitLabOAuthService {
	return &GitLabOAuthService{
		GitLabBaseURL: gitlabBaseURL,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BuildAuthURL builds the GitLab OAuth authorization URL
func (s *GitLabOAuthService) BuildAuthURL(state string) string {
	authURL := fmt.Sprintf("%s/oauth/authorize", s.GitLabBaseURL)

	params := url.Values{}
	params.Set("client_id", s.ClientID)
	params.Set("redirect_uri", s.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "read_user read_api")
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for an access token
func (s *GitLabOAuthService) ExchangeCode(ctx context.Context, code string) (*GitLabTokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/oauth/token", s.GitLabBaseURL)

	params := url.Values{}
	params.Set("client_id", s.ClientID)
	params.Set("client_secret", s.ClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", s.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab OAuth error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp GitLabTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUser fetches user information from GitLab using an access token
func (s *GitLabOAuthService) GetUser(ctx context.Context, accessToken string) (*GitLabUser, error) {
	userURL := fmt.Sprintf("%s/api/v4/user", s.GitLabBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var user GitLabUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *GitLabOAuthService) RefreshToken(ctx context.Context, refreshToken string) (*GitLabTokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/oauth/token", s.GitLabBaseURL)

	params := url.Values{}
	params.Set("client_id", s.ClientID)
	params.Set("client_secret", s.ClientSecret)
	params.Set("refresh_token", refreshToken)
	params.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab OAuth error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp GitLabTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}
