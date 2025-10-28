package websocket

import (
	"database/sql" // <-- ADICIONAR IMPORT
	"encoding/json"
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
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
	BacklogCount int64            `json:"backlogCount"` // Contagem de itens
	BacklogValue int64            `json:"backlogValue"` // Soma dos valores (P=250, M=80, G=10)
	BufferCounts map[string]int64 `json:"bufferCounts"` // Contagem por buffer
	BufferValues map[string]int64 `json:"bufferValues"` // Soma de valores por buffer
	BufferAvgTimes map[string]float64 `json:"bufferAvgTimes"` // Tempo médio (só RTS e EHA)
	
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
	// --- ASSINATURA DA FUNÇÃO MUDOU ---
	queue, backlogCount, backlogValue, bufferCounts, bufferValues, bufferAvgTimes := getCurrentQueueState()
	messageData := QueueUpdateMessage{
		Type:           "queue_update",
		Queue:          queue,
		BacklogCount:   backlogCount,
		BacklogValue:   backlogValue,
		BufferCounts:   bufferCounts,
		BufferValues:   bufferValues,
		BufferAvgTimes: bufferAvgTimes, // NOVO
	}
	// ... (resto da função inalterado) ...
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

func getCurrentQueueState() ([]models.Package, int64, int64, map[string]int64, map[string]int64, map[string]float64) {
	var packages []models.Package
	var backlogCount int64 = 0
	var backlogValue int64 = 0
	bufferCounts := make(map[string]int64)
	bufferValues := make(map[string]int64)
	bufferTotalSeconds := make(map[string]float64) // Para calcular a média
	bufferAvgTimes := make(map[string]float64)
	now := services.GetBrasiliaTime()

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		// Busca todos os pacotes ativos
		if err := tx.Where("buffer <> ?", "PENDENTE").Order("entry_timestamp asc").Find(&packages).Error; err != nil {
			return err
		}

		// Inicializa mapas
		bufferCounts["RTS"], bufferCounts["EHA"], bufferCounts["SAL"] = 0, 0, 0
		bufferValues["RTS"], bufferValues["EHA"], bufferValues["SAL"] = 0, 0, 0
		bufferTotalSeconds["RTS"], bufferTotalSeconds["EHA"] = 0.0, 0.0 // SAL não precisa
		bufferAvgTimes["RTS"], bufferAvgTimes["EHA"] = 0.0, 0.0         // SAL não terá

		for _, pkg := range packages {
			// Calcula tempo de permanência se o timestamp não for zero
			var durationSeconds float64 = 0
			if !pkg.EntryTimestamp.IsZero() {
				durationSeconds = now.Sub(pkg.EntryTimestamp).Seconds()
			}

			// Conta e Soma para BACKLOG (Excluindo SAL)
			if pkg.Buffer != "SAL" {
				backlogCount++
				backlogValue += int64(pkg.ProfileValue)
			}

			// Conta, Soma e Tempo Total por BUFFER
			switch pkg.Buffer {
			case "RTS":
				bufferCounts["RTS"]++
				bufferValues["RTS"] += int64(pkg.ProfileValue)
				bufferTotalSeconds["RTS"] += durationSeconds
			case "EHA":
				bufferCounts["EHA"]++
				bufferValues["EHA"] += int64(pkg.ProfileValue)
				bufferTotalSeconds["EHA"] += durationSeconds
			case "SAL":
				bufferCounts["SAL"]++
				// bufferValues["SAL"] permanece 0
				// bufferTotalSeconds["SAL"] não é calculado
			}
		}

		// Calcula Tempo Médio
		if bufferCounts["RTS"] > 0 {
			bufferAvgTimes["RTS"] = bufferTotalSeconds["RTS"] / float64(bufferCounts["RTS"])
		}
		if bufferCounts["EHA"] > 0 {
			bufferAvgTimes["EHA"] = bufferTotalSeconds["EHA"] / float64(bufferCounts["EHA"])
		}

		return nil
	}, &sql.TxOptions{ReadOnly: true})

	if err != nil {
		log.Printf("Erro ao buscar estado da fila no DB: %v", err)
		// Retorna zero para todos os valores em caso de erro
		emptyCounts := map[string]int64{"RTS": 0, "EHA": 0, "SAL": 0}
		emptyValues := map[string]int64{"RTS": 0, "EHA": 0, "SAL": 0}
		emptyAvgTimes := map[string]float64{"RTS": 0.0, "EHA": 0.0}
		return []models.Package{}, 0, 0, emptyCounts, emptyValues, emptyAvgTimes
	}

	return packages, backlogCount, backlogValue, bufferCounts, bufferValues, bufferAvgTimes
}

func (h *Hub) BroadcastQueueUpdate() {
	// --- ASSINATURA DA FUNÇÃO MUDOU ---
	queue, backlogCount, backlogValue, bufferCounts, bufferValues, bufferAvgTimes := getCurrentQueueState()
	messageData := QueueUpdateMessage{
		Type:           "queue_update",
		Queue:          queue,
		BacklogCount:   backlogCount,
		BacklogValue:   backlogValue,
		BufferCounts:   bufferCounts,
		BufferValues:   bufferValues,
		BufferAvgTimes: bufferAvgTimes, // NOVO
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