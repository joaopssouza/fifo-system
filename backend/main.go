// backend/main.go
package main

import (
	"fifo-system/backend/config"
	"fifo-system/backend/controllers"
	"fifo-system/backend/initializers"
	"fifo-system/backend/middleware"
	"fifo-system/backend/models"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadConfig()
	initializers.ConnectToDB()
}

func main() {
	log.Println("Iniciando a migração...")
	err := initializers.DB.AutoMigrate(&models.User{}, &models.Package{}, &models.AuditLog{})
	if err != nil {
		log.Fatal("Migração falhou:", err)
	}
	log.Println("Migração concluída com sucesso.")

	r := gin.Default()

	// CORS Configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"} // Frontend URL
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// --- Rotas Públicas ---
	r.POST("/login", controllers.Login)

	// --- Rotas Protegidas ---
	api := r.Group("/api")
	api.Use(middleware.RequireAuth) // Todas as rotas neste grupo exigem autenticação
	{
		// Rotas para visualização de dados (qualquer utilizador autenticado pode aceder)
		api.GET("/fifo-queue", controllers.GetFIFOQueue)
		api.GET("/backlog-count", controllers.GetBacklogCount)

		// --- NOVA ROTA ADICIONADA AQUI ---
		// Rota para o utilizador alterar a sua própria senha.
		// O método é PUT, pois estamos a atualizar um recurso existente (a senha do utilizador).
		api.PUT("/user/change-password", controllers.ChangePassword)
		// --- FIM DA ADIÇÃO ---

		// Rotas para ações (restritas às roles 'admin' ou 'fifo')
		actions := api.Group("/")
		actions.Use(middleware.RequireRole("admin", "fifo"))
		{
			actions.POST("/entry", controllers.PackageEntry)
			actions.POST("/exit", controllers.PackageExit)
		}

		// Rotas apenas para admin
		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/logs", controllers.GetAuditLogs)
		}
	}

	log.Println("Iniciando o servidor na porta 8080...")
	r.Run(":8080")
}
