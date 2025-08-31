// backend/controllers/userController.go
package controllers

import (
	"fifo-system/backend/config"
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser cria um novo utilizador (ação de administrador)
func CreateUser(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"` // Role agora é obrigatória
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer nome de utilizador, senha e função."})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar a senha."})
		return
	}

	// Set a default role if not provided
	if body.Role == "" {
		body.Role = "fifo"
	}

	// Create the user
	user := models.User{Username: body.Username, PasswordHash: string(hash), Role: body.Role}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o utilizador."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilizador criado com sucesso."})
}

// NOVA FUNÇÃO: Listar todos os utilizadores (sem as senhas)
func GetUsers(c *gin.Context) {
	var users []models.User
	// Omit("password_hash") garante que nunca enviamos as senhas para o frontend
	initializers.DB.Omit("password_hash").Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func Login(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer nome de utilizador e senha."})
		return
	}

	// Look up requested user
	var user models.User
	initializers.DB.First(&user, "username = ?", body.Username)

	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Nome de utilizador ou senha inválidos"})
		return
	}

	// Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Nome de utilizador ou senha inválidos"})
		return
	}

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"user": user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(), // Token expires in 30 days
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Send it back
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}
func ChangePassword(c *gin.Context) {
	// Obter o ID do utilizador a partir do token JWT (anexado pelo middleware)
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(models.User)

	var body struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer a senha antiga e a nova."})
		return
	}

	// 1. Verificar se a senha antiga está correta
	err := bcrypt.CompareHashAndPassword([]byte(currentUser.PasswordHash), []byte(body.OldPassword))
	if err != nil {
		// Se as senhas não corresponderem, retorna um erro de não autorizado
		c.JSON(http.StatusUnauthorized, gin.H{"error": "A senha antiga está incorreta."})
		return
	}

	// 2. Gerar o hash para a nova senha
	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar a nova senha."})
		return
	}

	// 3. Atualizar a senha no banco de dados
	result := initializers.DB.Model(&currentUser).Update("password_hash", string(newHash))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar a senha."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Senha alterada com sucesso."})
}

// AdminResetPassword permite que um admin redefina a senha de outro utilizador.
func AdminResetPassword(c *gin.Context) {
	adminUser, _ := c.Get("user")
	currentAdmin := adminUser.(models.User)

	targetUserID := c.Param("id")

	var body struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer a nova senha."})
		return
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer a nova senha."})
		return
	}

	var targetUser models.User
	if err := initializers.DB.First(&targetUser, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilizador alvo não encontrado."})
		return
	}

	// Regra de Segurança: Um admin não pode redefinir a sua própria senha através desta rota.
	if targetUser.ID == currentAdmin.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Não pode redefinir a sua própria senha aqui. Use a funcionalidade 'Alterar Senha'."})
		return
	}
	// Gerar o hash para a nova senha
	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar a nova senha."})
		return
	}

	result := initializers.DB.Model(&targetUser).Update("password_hash", string(newHash))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar a senha."})
		return
	}

	// A chamada para CreateAuditLog foi REMOVIDA, conforme a sua instrução.

	c.JSON(http.StatusOK, gin.H{"message": "Senha do utilizador redefinida com sucesso."})
}
