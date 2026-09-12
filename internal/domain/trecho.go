package domain

// Trecho representa o percurso entre duas cidades consecutivas de uma carona.
// A disponibilidade e controlada separadamente para cada trecho.
type Trecho struct {
	CaronaID            string
	Ordem               int
	Origem              string
	Destino             string
	Capacidade          int
	AssentosDisponiveis int
	PrecoCentavos       int64
}
