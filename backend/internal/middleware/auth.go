package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT payload structure.
type Claims struct {
	UserID    string   `json:"sub"`
	Email     string   `json:"email"`
	Role      string   `json:"role"`
	Username  string   `json:"username"`
	TokenType string   `json:"token_type"`
	ProjectID string   `json:"project_id"`
	Scopes    []string `json:"scopes"`
	jwt.RegisteredClaims
}

// JWT returns a Gin middleware that validates Bearer tokens.
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)
		c.Set("token_type", claims.TokenType)
		c.Set("project_id", claims.ProjectID)
		c.Set("scopes", claims.Scopes)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[strings.ToLower(strings.TrimSpace(r))] = true
	}
	return func(c *gin.Context) {
		role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
		if !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}
