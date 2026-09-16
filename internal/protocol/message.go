package protocol

import "encoding/json"

type Requisicao struct {
	Operacao string          `json:"operacao"`
	Dados    json.RawMessage `json:"dados"`
}

type Resposta struct {
	Sucesso  bool            `json:"sucesso"`
	Mensagem string          `json:"mensagem"`
	Dados    json.RawMessage `json:"dados,omitempty"`
}

type CriarCarona struct {
	ID             string   `json:"id"`
	MotoristaID    string   `json:"motorista_id"`
	HorarioSaida   string   `json:"horario_saida"`
	Rota           []string `json:"rota"`
	Capacidade     int      `json:"capacidade"`
	PrecosCentavos []int64  `json:"precos_centavos"`
}

type BuscarItinerarios struct {
	Origem             string `json:"origem"`
	Destino            string `json:"destino"`
	QuantidadeAssentos int    `json:"quantidade_assentos"`
}

type ItinerarioEncontrado struct {
	Origem             string             `json:"origem"`
	Destino            string             `json:"destino"`
	Trechos            []TrechoEncontrado `json:"trechos"`
	PrecoTotalCentavos int64              `json:"preco_total_centavos"`
}

type TrechoEncontrado struct {
	CaronaID            string `json:"carona_id"`
	Ordem               int    `json:"ordem"`
	Origem              string `json:"origem"`
	Destino             string `json:"destino"`
	PrecoCentavos       int64  `json:"preco_centavos"`
	AssentosDisponiveis int    `json:"assentos_disponiveis"`
}
