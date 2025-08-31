package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"fifo-system/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserID   uint
	Username string
	Role     string
	Sector   string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Permite todas as origens por agora. Em produção, restrinja isto.
		return true
	},
}

var H = Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected: %s", client.Username)
			h.broadcastOnlineUsers()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s", client.Username)
			h.broadcastOnlineUsers()
		}
	}
}

func (h *Hub) broadcastOnlineUsers() {
	h.mu.Lock()
	defer h.mu.Unlock()

	var onlineUsers []map[string]interface{}
	for client := range h.clients {
		onlineUsers = append(onlineUsers, map[string]interface{}{
			"id":       client.UserID,
			"username": client.Username,
			"role":     client.Role,
			"sector":   client.Sector,
		})
	}

	message, err := json.Marshal(map[string]interface{}{
		"type": "online_users",
		"data": onlineUsers,
	})

	if err != nil {
		log.Printf("Error marshaling online users: %v", err)
		return
	}

	for client := range h.clients {
		// Envia a lista de utilizadores online apenas para os administradores
		if client.Role == "admin" {
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Websocket write error: %s", err)
			}
		}
	}
}

func ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	userInterface, exists := c.Get("user")
	if !exists {
		log.Println("User not found in context for WebSocket")
		conn.Close()
		return
	}

	currentUser := userInterface.(models.User)

	client := &Client{
		Conn:     conn,
		UserID:   currentUser.ID,
		Username: currentUser.Username,
		Role:     currentUser.Role,
		Sector:   currentUser.Sector,
	}

	H.register <- client

	// Rotina para lidar com a desconexão do cliente
	go func() {
		defer func() {
			H.unregister <- client
		}()
		for {
			// Apenas lê mensagens para detetar a desconexão. Não fazemos nada com a mensagem.
			if _, _, err := client.Conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}