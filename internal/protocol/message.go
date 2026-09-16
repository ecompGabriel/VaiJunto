package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const VersaoAtual = "1.0"

const (
	PerfilMotorista  = "motorista"
	PerfilPassageiro = "passageiro"
)

type Requisicao struct {
	Versao   string          `json:"versao"`
	ID       string          `json:"id"`
	Operacao string          `json:"operacao"`
	Dados    json.RawMessage `json:"dados"`
}

type Resposta struct {
	Versao   string          `json:"versao"`
	ID       string          `json:"id"`
	Sucesso  bool            `json:"sucesso"`
	Codigo   string          `json:"codigo"`
	Mensagem string          `json:"mensagem"`
	Dados    json.RawMessage `json:"dados,omitempty"`
}

type RegistrarUsuario struct {
	UsuarioID string `json:"usuario_id"`
	Senha     string `json:"senha"`
	Perfil    string `json:"perfil"`
}

type IniciarSessao struct {
	UsuarioID string `json:"usuario_id"`
	Senha     string `json:"senha"`
}

type SessaoIniciada struct {
	UsuarioID string `json:"usuario_id"`
	Perfil    string `json:"perfil"`
}

type CriarCarona struct {
	ID             string   `json:"id"`
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

type ReferenciaTrecho struct {
	CaronaID string `json:"carona_id"`
	Ordem    int    `json:"ordem"`
}

type ConfirmarReserva struct {
	IDReserva          string             `json:"id_reserva"`
	QuantidadeAssentos int                `json:"quantidade_assentos"`
	Trechos            []ReferenciaTrecho `json:"trechos"`
}

type CancelarReserva struct {
	IDReserva string `json:"id_reserva"`
}

type Reserva struct {
	ID                 string             `json:"id"`
	PassageiroID       string             `json:"passageiro_id"`
	QuantidadeAssentos int                `json:"quantidade_assentos"`
	Status             string             `json:"status"`
	Trechos            []ReferenciaTrecho `json:"trechos"`
}

type PassageiroConfirmado struct {
	PassageiroID       string `json:"passageiro_id"`
	QuantidadeAssentos int    `json:"quantidade_assentos"`
}

type TrechoDoMotorista struct {
	Ordem               int                    `json:"ordem"`
	Origem              string                 `json:"origem"`
	Destino             string                 `json:"destino"`
	Capacidade          int                    `json:"capacidade"`
	AssentosDisponiveis int                    `json:"assentos_disponiveis"`
	Passageiros         []PassageiroConfirmado `json:"passageiros"`
}

type CaronaDoMotorista struct {
	ID      string              `json:"id"`
	Trechos []TrechoDoMotorista `json:"trechos"`
}

func DecodificarEstrito(dados []byte, destino any) error {
	decodificador := json.NewDecoder(bytes.NewReader(dados))
	decodificador.DisallowUnknownFields()

	err := decodificador.Decode(destino)
	if err != nil {
		return err
	}

	err = decodificador.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("a mensagem contém dados JSON adicionais")
	}

	return nil
}
