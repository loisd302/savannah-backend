package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// OIDCConfig holds OIDC configuration
type OIDCConfig struct {
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	VerifyOptions *oidc.Config
}

// OIDCProvider handles OIDC authentication
type OIDCProvider struct {
	config   *OIDCConfig
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth2   oauth2.Config
	mutex    sync.RWMutex
}

// Claims represents JWT claims with custom fields
type Claims struct {
	jwt.RegisteredClaims
	Scope       string   `json:"scope,omitempty"`
	Scopes      []string `json:"scp,omitempty"`        // Azure AD format
	Permissions []string `json:"permissions,omitempty"` // Auth0 format
	Roles       []string `json:"roles,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
	Username    string   `json:"preferred_username,omitempty"`
}

// UserInfo contains authenticated user information
type UserInfo struct {
	Subject     string   `json:"sub"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Scopes      []string `json:"scopes"`
	Roles       []string `json:"roles"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider(config *OIDCConfig) (*OIDCProvider, error) {
	if config.VerifyOptions == nil {
		config.VerifyOptions = &oidc.Config{
			ClientID: config.ClientID,
		}
	}

	provider := &OIDCProvider{
		config: config,
	}

	if err := provider.initialize(); err != nil {
		return nil, err
	}

	return provider, nil
}

// initialize sets up the OIDC provider and verifier
func (p *OIDCProvider) initialize() error {
	ctx := context.Background()

	// Discover OIDC provider
	provider, err := oidc.NewProvider(ctx, p.config.IssuerURL)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	p.provider = provider
	p.verifier = provider.Verifier(p.config.VerifyOptions)

	// Setup OAuth2 config
	p.oauth2 = oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		RedirectURL:  p.config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       append([]string{oidc.ScopeOpenID}, p.config.Scopes...),
	}

	log.Printf("OIDC provider initialized for issuer: %s", p.config.IssuerURL)
	return nil
}

// AuthURL returns the OAuth2 authorization URL
func (p *OIDCProvider) AuthURL(state string) string {
	return p.oauth2.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for tokens
func (p *OIDCProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth2.Exchange(ctx, code)
}

// VerifyToken verifies an ID token and returns claims
func (p *OIDCProvider) VerifyToken(ctx context.Context, rawToken string) (*Claims, error) {
	// Verify with OIDC provider
	idToken, err := p.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Parse claims
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// VerifyAccessToken verifies an access token (JWT format)
func (p *OIDCProvider) VerifyAccessToken(tokenString string) (*Claims, error) {
	// For now, use a simpler approach - this would need proper JWKS handling
	// In production, you would fetch and cache JWKS from the provider's .well-known endpoint
	// and validate the JWT signature properly
	_ = tokenString // Avoid unused parameter warning
	return nil, fmt.Errorf("access token validation not fully implemented - use ID token validation instead")
}

// ExtractUserInfo extracts user information from claims
func (p *OIDCProvider) ExtractUserInfo(claims *Claims) *UserInfo {
	userInfo := &UserInfo{
		Subject:  claims.Subject,
		Email:    claims.Email,
		Name:     claims.Name,
		Username: claims.Username,
		Roles:    claims.Roles,
		Groups:   claims.Groups,
	}

	// Extract scopes from different claim formats
	if claims.Scope != "" {
		userInfo.Scopes = strings.Split(claims.Scope, " ")
	}
	if len(claims.Scopes) > 0 {
		userInfo.Scopes = claims.Scopes
	}
	if len(claims.Permissions) > 0 {
		userInfo.Permissions = claims.Permissions
	}

	return userInfo
}

// RequireAuth middleware that requires valid authentication
func (p *OIDCProvider) RequireAuth() gin.HandlerFunc {
	return p.RequireScopes() // No specific scopes required, just authentication
}

// RequireScopes middleware that requires specific scopes
func (p *OIDCProvider) RequireScopes(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify access token
		claims, err := p.VerifyAccessToken(tokenString)
		if err != nil {
			log.Printf("Token verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check token expiry
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "token expired",
			})
			c.Abort()
			return
		}

		// Extract user info
		userInfo := p.ExtractUserInfo(claims)

		// Check required scopes
		if len(requiredScopes) > 0 && !p.hasRequiredScopes(userInfo, requiredScopes) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient scope",
				"required_scopes": requiredScopes,
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user", userInfo)
		c.Set("claims", claims)
		c.Next()
	}
}

// RequireRoles middleware that requires specific roles
func (p *OIDCProvider) RequireRoles(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure authentication
		p.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userInfo, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "user info not found in context",
			})
			c.Abort()
			return
		}

		user := userInfo.(*UserInfo)
		if !p.hasRequiredRoles(user, requiredRoles) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient privileges",
				"required_roles": requiredRoles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to check if user has required scopes
func (p *OIDCProvider) hasRequiredScopes(user *UserInfo, requiredScopes []string) bool {
	userScopeMap := make(map[string]bool)
	for _, scope := range user.Scopes {
		userScopeMap[scope] = true
	}
	for _, scope := range user.Permissions {
		userScopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if !userScopeMap[required] {
			return false
		}
	}
	return true
}

// Helper function to check if user has required roles
func (p *OIDCProvider) hasRequiredRoles(user *UserInfo, requiredRoles []string) bool {
	userRoleMap := make(map[string]bool)
	for _, role := range user.Roles {
		userRoleMap[role] = true
	}

	for _, required := range requiredRoles {
		if !userRoleMap[required] {
			return false
		}
	}
	return true
}

// GetCurrentUser returns the current authenticated user from context
func GetCurrentUser(c *gin.Context) (*UserInfo, bool) {
	if user, exists := c.Get("user"); exists {
		return user.(*UserInfo), true
	}
	return nil, false
}