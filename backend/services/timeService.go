// backend/services/timeService.go
package services

import (
	"encoding/json"
	"net/http"
	"time"
)

// Estrutura para corresponder à resposta da API WorldTime
type WorldTimeResponse struct {
	UTCDateTime string `json:"utc_datetime"`
}

// GetWorldTime busca a hora UTC atual de uma API externa.
func GetWorldTime() (time.Time, error) {
	resp, err := http.Get("http://worldtimeapi.org/api/timezone/Etc/UTC")
	if err != nil {
		// Se a API externa falhar, usamos o tempo do servidor como fallback
		// para manter o sistema a funcionar, mas registamos o erro.
		return time.Now().UTC(), err
	}
	defer resp.Body.Close()

	var worldTime WorldTimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&worldTime); err != nil {
		return time.Now().UTC(), err
	}

	// Faz o parse da data retornada pela API
	parsedTime, err := time.Parse(time.RFC3339, worldTime.UTCDateTime)
	if err != nil {
		return time.Now().UTC(), err
	}

	return parsedTime, nil
}