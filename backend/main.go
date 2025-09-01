// backend/main.go
package main

import (
	"fifo-system/backend/config"
	"fifo-system/backend/controllers"
	"fifo-system/backend/initializers"
	"fifo-system/backend/middleware"
	"fifo-system/backend/models"
	"fifo-system/backend/websocket"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	config.LoadConfig()
	initializers.ConnectToDB()
}

func main() {
	log.Println("Iniciando a migração da base de dados...")
	// Adiciona a tabela de junção explícita à migração
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

	r.POST("/login", controllers.Login)

	api := r.Group("/api")
	api.Use(middleware.RequireAuth)
	{
		api.GET("/ws", websocket.ServeWs)
		api.PUT("/user/change-password", controllers.ChangePassword)
		api.GET("/fifo-queue", controllers.GetFIFOQueue)
		api.GET("/backlog-count", controllers.GetBacklogCount)
		api.POST("/entry", middleware.RequirePermission("MANAGE_FIFO"), controllers.PackageEntry)
		api.POST("/exit", middleware.RequirePermission("MANAGE_FIFO"), controllers.PackageExit)

		management := r.Group("/api/management")
		management.Use(middleware.RequireAuth)
		{
			management.GET("/roles", middleware.RequirePermission("EDIT_USER"), controllers.GetRoles)
			management.POST("/users", middleware.RequirePermission("CREATE_USER"), controllers.CreateUser)
			management.GET("/users", middleware.RequirePermission("VIEW_USERS"), controllers.GetUsers)
			management.PUT("/users/:id", middleware.RequirePermission("EDIT_USER"), controllers.AdminUpdateUser)
			management.PUT("/users/:id/reset-password", middleware.RequirePermission("RESET_PASSWORD"), controllers.AdminResetPassword)
			management.GET("/logs", middleware.RequirePermission("VIEW_LOGS"), controllers.GetAuditLogs)
		}
	}

	log.Println("Iniciando o servidor na porta 8080...")
	r.Run(":8080")
}

func seedAdminUser() {
	var userCount int64
	initializers.DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		var adminRole models.Role
		if err := initializers.DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			log.Fatalf("Erro ao semear utilizador admin: papel 'admin' não encontrado. Certifique-se de que seedData() é executado primeiro.")
			return
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 10)
		admin := models.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Sector:       "Administração",
			RoleID:       adminRole.ID,
		}
		initializers.DB.Create(&admin)
		log.Println("Utilizador 'admin' inicial criado com sucesso.")
	}
}

// --- VERSÃO DEFINITIVA E À PROVA DE FALHAS DA FUNÇÃO SEEDDATA ---
func seedData() {
	var roleCount int64
	initializers.DB.Model(&models.Role{}).Count(&roleCount)
	if roleCount > 0 {
		log.Println("Os dados de papéis e permissões já existem. A saltar o seeding.")
		return
	}

	log.Println("A semear dados iniciais de Papéis e Permissões...")

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Criar todas as permissões primeiro
		permissions := []models.Permission{
			{Name: "MANAGE_FIFO", Description: "Pode realizar entradas e saídas na fila"},
			{Name: "VIEW_LOGS", Description: "Pode visualizar os logs de atividade"},
			{Name: "VIEW_USERS", Description: "Pode visualizar a lista de utilizadores"},
			{Name: "CREATE_USER", Description: "Pode criar novos utilizadores"},
			{Name: "EDIT_USER", Description: "Pode editar o papel e setor de outros utilizadores"},
			{Name: "RESET_PASSWORD", Description: "Pode redefinir a senha de outros utilizadores"},
		}
		if err := tx.Create(&permissions).Error; err != nil {
			return err
		}
		log.Println("Permissões criadas com sucesso.")

		// 2. Criar os papéis (ainda sem associações)
		roles := []models.Role{
			{Name: "admin", Description: "Administrador do Sistema"},
			{Name: "leader", Description: "Líder de Equipa"},
			{Name: "fifo", Description: "Operador FIFO"},
		}
		if err := tx.Create(&roles).Error; err != nil {
			return err
		}
		log.Println("Papéis criados com sucesso.")

		// 3. AGORA, com os papéis e permissões já existentes, criamos as associações
		log.Println("A associar permissões aos papéis...")

		// Associação para Admin (todas as permissões)
		var adminRole models.Role
		tx.First(&adminRole, "name = ?", "admin")
		if err := tx.Model(&adminRole).Association("Permissions").Append(&permissions); err != nil {
			return err
		}
		log.Println("Permissões de Admin associadas.")

		// Associação para Leader
		var leaderRole models.Role
		var leaderPermissions []models.Permission
		tx.Where("name IN ?", []string{"MANAGE_FIFO", "VIEW_LOGS", "VIEW_USERS", "EDIT_USER", "RESET_PASSWORD", "CREATE_USER"}).Find(&leaderPermissions)
		tx.First(&leaderRole, "name = ?", "leader")
		if err := tx.Model(&leaderRole).Association("Permissions").Append(&leaderPermissions); err != nil {
			return err
		}
		log.Println("Permissões de Leader associadas.")

		// Associação para FIFO
		var fifoRole models.Role
		var fifoPermission models.Permission
		tx.Where("name = ?", "MANAGE_FIFO").First(&fifoPermission)
		tx.First(&fifoRole, "name = ?", "fifo")
		if err := tx.Model(&fifoRole).Association("Permissions").Append(&fifoPermission); err != nil {
			return err
		}
		log.Println("Permissões de FIFO associadas.")

		return nil // Commit da transação
	})

	if err != nil {
		log.Fatalf("Falha ao semear a base de dados: %v", err)
	}

	log.Println("Seeding da base de dados concluído com sucesso.")
}
