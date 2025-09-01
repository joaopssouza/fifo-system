package websocket

import (
	"encoding/json"
	"fifo-system/backend/models"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Client representa um utilizador conectado via WebSocket.
type Client struct {
	Conn     *websocket.Conn
	UserID   uint
	FullName string
	Username string
	Role     string
	Sector   string
}

// Hub mantém o conjunto de clientes ativos.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Para desenvolvimento, permite todas as origens.
		// Em produção, deve restringir isto ao seu domínio de frontend.
		return true
	},
}

// H é a instância global do nosso Hub.
var H = Hub{
	clients:    make(map[*Client]bool),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

// Run inicia o processamento de eventos do Hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Cliente conectado: %s", client.Username)
			h.broadcastOnlineUsers()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mu.Unlock()
			log.Printf("Cliente desconectado: %s", client.Username)
			h.broadcastOnlineUsers()
		}
	}
}

// broadcastOnlineUsers envia a lista de utilizadores online para todos os admins e leaders conectados.
func (h *Hub) broadcastOnlineUsers() {
	h.mu.Lock()
	defer h.mu.Unlock()

	var onlineUsers []map[string]interface{}
	for client := range h.clients {
		onlineUsers = append(onlineUsers, map[string]interface{}{
			"fullName": client.FullName,
			"id":       client.UserID,
			"username": client.Username,
			"role":     client.Role,
			"sector":   client.Sector,
		})
	}

	message, _ := json.Marshal(map[string]interface{}{
		"type": "online_users",
		"data": onlineUsers,
	})

	for client := range h.clients {
		// --- LÓGICA ATUALIZADA ---
		// Agora, tanto 'admin' quanto 'leader' recebem a lista.
		if client.Role == "admin" || client.Role == "leader" {
			client.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// ServeWs lida com o upgrade de requisições HTTP para WebSocket.
func ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	userInterface, _ := c.Get("user")
	currentUser := userInterface.(models.User)

	client := &Client{
		Conn:     conn,
		UserID:   currentUser.ID,
		FullName: currentUser.FullName,
		Username: currentUser.Username,
		Role:     currentUser.Role.Name,
		Sector:   currentUser.Sector,
	}

	H.register <- client

	// Esta rotina escuta por mensagens para detetar quando o cliente se desconecta.
	go func() {
		defer func() {
			H.unregister <- client
		}()
		for {
			if _, _, err := client.Conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
