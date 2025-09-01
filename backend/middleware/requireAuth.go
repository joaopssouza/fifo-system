// backend/middleware/requireAuth.go
package middleware

import (
	"fifo-system/backend/config"
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	var tokenString string

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		tokenString = c.Query("token")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização não encontrado"})
			return
		}
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "O token expirou"})
			return
		}

		var user models.User
		// --- CORREÇÃO CRÍTICA ---
		// Recarrega sempre os dados do utilizador a partir do banco de dados para garantir que estão atualizados,
		// especialmente após migrações ou alterações de dados.
		if err := initializers.DB.Preload("Role.Permissions").First(&user, claims["sub"]).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Utilizador não encontrado"})
			return
		}

		c.Set("user", user)
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Claims do token inválidas"})
	}
}

// RequirePermission verifica se o usuário autenticado tem uma permissão específica.
func RequirePermission(permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Utilizador não encontrado no contexto"})
			return
		}

		user := userInterface.(models.User)
		hasPermission := false

		for _, p := range user.Role.Permissions {
			if p.Name == permissionName {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permissões insuficientes para aceder a este recurso."})
			return
		}

		c.Next()
	}
}
