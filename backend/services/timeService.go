// backend/services/timeService.go
package services

import (
	"log"
	"time"
)

var brasiliaLocation *time.Location

func init() {
	// Carrega a localização do fuso horário de São Paulo (que rege o horário de Brasília)
	// uma única vez quando a aplicação inicia, para máxima performance.
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatalf("FATAL: Falha ao carregar o fuso horário 'America/Sao_Paulo': %v", err)
	}
	brasiliaLocation = loc
}

// GetBrasiliaTime retorna a hora atual, sempre no fuso horário de Brasília.
// Esta função é agora a fonte oficial de tempo para todas as operações do sistema.
func GetBrasiliaTime() time.Time {
	return time.Now().In(brasiliaLocation)
}