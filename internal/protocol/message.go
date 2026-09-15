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
