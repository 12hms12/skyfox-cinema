package security

import (
	"slices"
	"net/http"
	"skyfox/common/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
    ctxUserIDKey = "userID"
    ctxRoleKey   = "role"
)

func JWTAuth(jwtManager *JwtManager, allowedRoles ...Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
            return
        }

        tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

        claims, err := jwtManager.ParseToken(tokenStr)
        if err != nil {
            logger.Error("invalid token: %v", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            return
        }

        if len(allowedRoles) > 0 && ! slices.Contains(allowedRoles, claims.Role) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            return
        }

        c.Set(ctxUserIDKey, claims.UserID)
        c.Set(ctxRoleKey, claims.Role)

        c.Next()
    }
}