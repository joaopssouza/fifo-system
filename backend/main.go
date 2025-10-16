// backend/main.go
package main

import (
	"fifo-system/backend/config"
	"fifo-system/backend/controllers"
	"fifo-system/backend/initializers"
	"fifo-system/backend/middleware"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
	"fifo-system/backend/websocket"
	"log"
	"net/http"
	"time"
	"os"
	
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	config.LoadConfig()
	initializers.ConnectToDB()
}

func main() {
	log.Println("Iniciando a migração da base de dados...")
	err := initializers.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.Package{}, &models.AuditLog{})
	if err != nil {
		log.Fatalf("Falha na migração da base de dados: %v", err)
	}

	seedData()
	seedAdminUser()

	go websocket.H.Run()
	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// --- ROTAS PÚBLICAS ---
	// Estas rotas NÃO passam pelo middleware de autenticação.
	r.POST("/login", controllers.Login)
	r.GET("/public/time", func(c *gin.Context) {
		serverTime, err := services.GetWorldTime()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"serverTime": time.Now().UTC()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"serverTime": serverTime})
	})
	r.GET("/public/fifo-queue", controllers.GetFIFOQueue)
	r.GET("/public/backlog-count", controllers.GetBacklogCount)

	// --- ROTAS PRIVADAS / PROTEGIDAS ---
	// Todas as rotas dentro deste grupo "/api" exigirão um token de autenticação válido.
	api := r.Group("/api")
	api.Use(middleware.RequireAuth)
	{
		api.GET("/ws", websocket.ServeWs)
		api.PUT("/user/change-password", controllers.ChangePassword)
		api.GET("/fifo-queue", controllers.GetFIFOQueue)
		api.GET("/backlog-count", controllers.GetBacklogCount)
		api.POST("/entry", middleware.RequirePermission("MANAGE_FIFO"), controllers.PackageEntry)
		api.POST("/exit", middleware.RequirePermission("MANAGE_FIFO"), controllers.PackageExit)
		api.PUT("/package/move/:id", middleware.RequirePermission("MOVE_PACKAGE"), controllers.MovePackage)
		api.POST("/qrcodes/generate-data", middleware.RequirePermission("GENERATE_QR_CODES"), controllers.GenerateQRCodeData)
		api.POST("/qrcodes/confirm", middleware.RequirePermission("GENERATE_QR_CODES"), controllers.ConfirmQRCodeData)
		api.GET("/qrcodes/find/:trackingId", middleware.RequirePermission("GENERATE_QR_CODES"), controllers.FindQRCodeData)

		management := api.Group("/management") // Corrigido para 'api.Group' para herdar a autenticação
		{
			management.GET("/roles", middleware.RequirePermission("EDIT_USER"), controllers.GetRoles)
			management.POST("/users", middleware.RequirePermission("CREATE_USER"), controllers.CreateUser)
			management.GET("/users", middleware.RequirePermission("VIEW_USERS"), controllers.GetUsers)
			management.PUT("/users/:id", middleware.RequirePermission("EDIT_USER"), controllers.AdminUpdateUser)
			management.PUT("/users/:id/reset-password", middleware.RequirePermission("RESET_PASSWORD"), controllers.AdminResetPassword)
			management.GET("/logs", middleware.RequirePermission("VIEW_LOGS"), controllers.GetAuditLogs)
		}
	}

	port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Mantém 8080 como padrão para desenvolvimento local
    }

    log.Printf("Iniciando o servidor na porta %s...", port)
    r.Run(":" + port) // <-- ALTERAÇÃO AQUI
}

// ... (funções seedAdminUser e seedData permanecem as mesmas)
func seedAdminUser() {
	var userCount int64
	initializers.DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		var adminRole models.Role
		if err := initializers.DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			log.Fatalf("Erro ao semear utilizador admin: papel 'admin' não encontrado...")
			return
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 10)
		admin := models.User{
			Username: "admin",
			FullName:     "Administrador do Sistema",
			PasswordHash: string(hash),
			Sector:       "ADMINISTRAÇÃO", 
			RoleID:       adminRole.ID,
		}
		initializers.DB.Create(&admin)
		log.Println("Utilizador 'admin' inicial criado com sucesso.")
	}
}

func seedData() {
	log.Println("Sincronizando papéis e permissões...")

	allPermissions := []models.Permission{
		{Name: "MANAGE_FIFO", Description: "Pode realizar entradas e saídas na fila"},
		{Name: "VIEW_LOGS", Description: "Pode visualizar os logs de atividade"},
		{Name: "VIEW_USERS", Description: "Pode visualizar a lista de utilizadores"},
		{Name: "CREATE_USER", Description: "Pode criar novos utilizadores"},
		{Name: "EDIT_USER", Description: "Pode editar o papel e setor de outros utilizadores"},
		{Name: "RESET_PASSWORD", Description: "Pode redefinir a senha de outros utilizadores"},
		{Name: "MOVE_PACKAGE", Description: "Pode mover um item para uma nova rua"},
		{Name: "GENERATE_QR_CODES", Description: "Pode gerar novos QR Codes de rastreamento"},
	}

	for _, p := range allPermissions {
		initializers.DB.FirstOrCreate(&p, models.Permission{Name: p.Name})
	}
	log.Println("Todas as permissões foram criadas ou verificadas.")

	rolesToPermissions := map[string][]string{
		"admin": {
			"MANAGE_FIFO", "VIEW_LOGS", "VIEW_USERS", "CREATE_USER",
			"EDIT_USER", "RESET_PASSWORD", "MOVE_PACKAGE", "GENERATE_QR_CODES",
		},
		"leader": {
			"MANAGE_FIFO", "VIEW_LOGS", "VIEW_USERS", "CREATE_USER",
			"EDIT_USER", "RESET_PASSWORD", "MOVE_PACKAGE", "GENERATE_QR_CODES",
		},
		"fifo": {
			"MANAGE_FIFO", "MOVE_PACKAGE",
		},
	}

	for roleName, permissionNames := range rolesToPermissions {
		var role models.Role
		initializers.DB.FirstOrCreate(&role, models.Role{Name: roleName})

		var permissionsToAssign []models.Permission
		initializers.DB.Where("name IN ?", permissionNames).Find(&permissionsToAssign)

		err := initializers.DB.Model(&role).Association("Permissions").Replace(&permissionsToAssign)
		if err != nil {
			log.Printf("Falha ao associar permissões para o papel '%s': %v\n", roleName, err)
		}
	}
	log.Println("Sincronização de papéis e permissões concluída com sucesso.")
}