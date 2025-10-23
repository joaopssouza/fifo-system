package websocket

import (
	"database/sql" // <-- ADICIONAR IMPORT
	"encoding/json"
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
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

// --- Estrutura para a mensagem de atualização da fila ---
type QueueUpdateMessage struct {
	Type         string           `json:"type"`
	Queue        []models.Package `json:"queue"`
	Backlog      int64            `json:"backlog"`
	BufferCounts map[string]int64 `json:"bufferCounts"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var H = Hub{
	clients:    make(map[*Client]bool),
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
			log.Printf("Cliente conectado: %s", client.Username)
			h.broadcastOnlineUsers()
			h.sendInitialQueueState(client)

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

func (h *Hub) sendInitialQueueState(client *Client) {
	queue, backlog, bufferCounts := getCurrentQueueState()
	messageData := QueueUpdateMessage{
		Type:         "queue_update",
		Queue:        queue,
		Backlog:      backlog,
		BufferCounts: bufferCounts,
	}

	message, err := json.Marshal(messageData)
	if err != nil {
		log.Printf("Erro ao serializar estado inicial da fila: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; ok {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Erro ao enviar estado inicial da fila para %s: %v", client.Username, err)
		}
	}
}

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
		if client.Role == "admin" || client.Role == "leader" {
			err := client.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Erro ao enviar lista de utilizadores online para %s: %v", client.Username, err)
			}
		}
	}
}

func getCurrentQueueState() ([]models.Package, int64, map[string]int64) {
	var packages []models.Package
	var count int64
	bufferCounts := make(map[string]int64)

	// --- CORREÇÃO: Usar *sql.TxOptions ---
	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("buffer <> ?", "PENDENTE").Order("entry_timestamp asc").Find(&packages).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Package{}).Where("buffer <> ?", "PENDENTE").Count(&count).Error; err != nil {
			return err
		}
		var rtsCount, ehaCount, salCount int64
		tx.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "RTS").Count(&rtsCount) // Adicionado IS NULL
		tx.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "EHA").Count(&ehaCount) // Adicionado IS NULL
		tx.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "SAL").Count(&salCount) // Adicionado IS NULL
		bufferCounts["RTS"] = rtsCount
		bufferCounts["EHA"] = ehaCount
		bufferCounts["SAL"] = salCount
		return nil
	}, &sql.TxOptions{ReadOnly: true}) // <-- CORRIGIDO AQUI
	// --- FIM DA CORREÇÃO ---

	if err != nil {
		log.Printf("Erro ao buscar estado da fila no DB: %v", err)
		return []models.Package{}, 0, map[string]int64{"RTS": 0, "EHA": 0, "SAL": 0}
	}

	return packages, count, bufferCounts
}

func (h *Hub) BroadcastQueueUpdate() {
	queue, backlog, bufferCounts := getCurrentQueueState()
	messageData := QueueUpdateMessage{
		Type:         "queue_update",
		Queue:        queue,
		Backlog:      backlog,
		BufferCounts: bufferCounts,
	}

	message, err := json.Marshal(messageData)
	if err != nil {
		log.Printf("Erro ao serializar atualização da fila: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Erro ao enviar atualização da fila para %s: %v", client.Username, err)
		}
	}
	log.Println("Atualização da fila enviada para todos os clientes.")
}

func ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Falha no upgrade para WebSocket: %v", err)
		return
	}

	userInterface, exists := c.Get("user")
	if !exists {
		log.Println("Tentativa de conexão WS sem utilizador autenticado.")
		conn.Close()
		return
	}
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

	go func() {
		defer func() {
			H.unregister <- client
			client.Conn.Close()
		}()
		client.Conn.SetReadLimit(maxMessageSize)
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		client.Conn.SetPongHandler(func(string) error { client.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

		for {
			_, _, err := client.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Erro de leitura WebSocket (cliente %s): %v", client.Username, err)
				}
				break
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
		}()
		for range ticker.C {
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Erro ao enviar ping para %s: %v", client.Username, err)
				return
			}
		}
	}()
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)