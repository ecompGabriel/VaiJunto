package domain

import "time"

type Carona struct {
	ID           string
	motoristaID  string
	horarioSaida time.Time
	rota         []string
	trechos      []Trecho
}
