// backend/controllers/userController.go
package controllers

import (
	"errors"
	"fifo-system/backend/config"
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser cria um novo utilizador (ação de administrador)
func CreateUser(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		FullName string `json:"fullName" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
		Sector   string `json:"sector" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todos os campos são obrigatórios."})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar a senha."})
		return
	}

	var role models.Role
	if err := initializers.DB.Where("name = ?", body.Role).First(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O papel especificado não existe."})
		return
	}

	user := models.User{
		Username:     body.Username,
		FullName:     body.FullName,
		PasswordHash: string(hash),
		Sector:       body.Sector,
		RoleID:       role.ID,
	}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "O nome de utilizador já existe."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o utilizador."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilizador criado com sucesso."})
}

// GetUsers lista todos os utilizadores, incluindo a informação do seu papel.
func GetUsers(c *gin.Context) {
	var users []models.User
	// --- CORREÇÃO CRÍTICA ---
	// Usamos Preload("Role") para carregar os dados do papel associado a cada utilizador.
	// Omit("password_hash") garante que nunca enviamos as senhas para o frontend.
	if err := initializers.DB.Preload("Role").Omit("password_hash").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar utilizadores."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// Login gera um token JWT para um utilizador autenticado.
func Login(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer nome de utilizador e senha."})
		return
	}

	var user models.User
	// --- INÍCIO DA CORREÇÃO ---
	// Carrega o utilizador e todas as suas associações (Role e Permissions) de uma só vez.
	if err := initializers.DB.Preload("Role.Permissions").First(&user, "username = ?", body.Username).Error; err != nil {
		// Se o utilizador não for encontrado, retornamos um erro genérico para segurança.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Nome de utilizador ou senha inválidos"})
		return
	}
	// --- FIM DA CORREÇÃO ---

	// Compara a senha fornecida com o hash armazenado.
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Nome de utilizador ou senha inválidos"})
		return
	}

	// Extrai os nomes das permissões para incluir no token.
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// Log de depuração para verificar as permissões.
	log.Printf("A gerar token para Utilizador: '%s' com as permissões: %v", user.Username, permissions)

	// Cria o token JWT com todos os dados corretos e consistentes.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         user.ID,
		"user":        user.Username,
		"fullName":    user.FullName, // Agora o FullName está sempre presente.
		"role":        user.Role.Name,
		"permissions": permissions,
		"exp":         time.Now().Add(config.AppConfig.JWTExpirationTime).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// ChangePassword permite que um utilizador autenticado altere a sua própria senha.
func ChangePassword(c *gin.Context) {
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

	err := bcrypt.CompareHashAndPassword([]byte(currentUser.PasswordHash), []byte(body.OldPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "A senha antiga está incorreta."})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar a nova senha."})
		return
	}

	result := initializers.DB.Model(&currentUser).Update("password_hash", string(newHash))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar a senha."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Senha alterada com sucesso."})
}

// GetRoles lista todos os papéis disponíveis no sistema.
func GetRoles(c *gin.Context) {
	var roles []models.Role
	initializers.DB.Find(&roles)
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// validateAdminAction executa as verificações de hierarquia para ações de administração.
func validateAdminAction(actingUser models.User, targetUserID uint) (*models.User, error) {
	// Regra 1: Um utilizador não pode editar a si mesmo.
	if actingUser.ID == targetUserID {
		return nil, errors.New("não pode executar esta ação no seu próprio perfil")
	}

	var targetUser models.User
	if err := initializers.DB.Preload("Role").First(&targetUser, targetUserID).Error; err != nil {
		return nil, errors.New("utilizador alvo não encontrado")
	}

	// Regra 2: Ninguém pode editar um 'admin'.
	if targetUser.Role.Name == "admin" {
		return nil, errors.New("não é permitido modificar um administrador")
	}

	// Regra 3: Um 'leader' não pode editar outro 'leader'.
	if actingUser.Role.Name == "leader" && targetUser.Role.Name == "leader" {
		return nil, errors.New("líderes não podem editar outros líderes")
	}

	return &targetUser, nil
}

// AdminUpdateUser permite que um administrador atualize o papel e setor de outro utilizador.
func AdminUpdateUser(c *gin.Context) {
	actingUserInterface, _ := c.Get("user")
	actingUser := actingUserInterface.(models.User)

	targetUserIDStr := c.Param("id")
	targetUserID, _ := strconv.ParseUint(targetUserIDStr, 10, 32)

	// --- LÓGICA DE VALIDAÇÃO REATORIZADA ---
	targetUser, err := validateAdminAction(actingUser, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	// --- FIM DA REATORIZAÇÃO ---

	var body struct {
		FullName string `json:"fullName"`
		RoleID   uint   `json:"roleId"`
		Sector   string `json:"sector"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos."})
		return
	}

	updates := models.User{FullName: body.FullName, RoleID: body.RoleID, Sector: body.Sector}
	initializers.DB.Model(targetUser).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "Utilizador atualizado com sucesso."})
}

// AdminResetPassword (aplicar a mesma lógica de validação)
func AdminResetPassword(c *gin.Context) {
	actingUserInterface, _ := c.Get("user")
	actingUser := actingUserInterface.(models.User)

	targetUserIDStr := c.Param("id")
	targetUserID, _ := strconv.ParseUint(targetUserIDStr, 10, 32)

	// --- LÓGICA DE VALIDAÇÃO REATORIZADA ---
	targetUser, err := validateAdminAction(actingUser, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	// --- FIM DA REATORIZAÇÃO ---

	var body struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário fornecer a nova senha."})
		return
	}
	newHash, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	initializers.DB.Model(targetUser).Update("password_hash", string(newHash))
	c.JSON(http.StatusOK, gin.H{"message": "Senha do utilizador redefinida com sucesso."})
}
