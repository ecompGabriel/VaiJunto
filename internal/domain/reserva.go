package domain

import (
	"errors"
	"strings"
)

type StatusReserva string

const (
	StatusReservaConfirmada StatusReserva = "confirmada"
	StatusReservaCancelada  StatusReserva = "cancelada"
)

type ReferenciaTrecho struct {
	CaronaID string
	Ordem    int
}

type Reserva struct {
	ID                 string
	PassageiroID       string
	QuantidadeAssentos int
	Status             StatusReserva
	trechos            []ReferenciaTrecho
}

func NovaReserva(
	id string,
	passageiroID string,
	quantidadeAssentos int,
	trechos []ReferenciaTrecho,
) (*Reserva, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("o ID da reserva é obrigatório")
	}

	if strings.TrimSpace(passageiroID) == "" {
		return nil, errors.New("o ID do passageiro é obrigatório")
	}

	if quantidadeAssentos <= 0 {
		return nil, errors.New("a quantidade de assentos deve ser positiva")
	}

	if len(trechos) == 0 {
		return nil, errors.New("a reserva precisa ter ao menos um trecho")
	}

	referenciasVistas := make(map[ReferenciaTrecho]bool)

	for _, trecho := range trechos {
		if strings.TrimSpace(trecho.CaronaID) == "" {
			return nil, errors.New("o ID da carona do trecho é obrigatório")
		}

		if trecho.Ordem < 0 {
			return nil, errors.New("a ordem do trecho não pode ser negativa")
		}

		if referenciasVistas[trecho] {
			return nil, errors.New("um trecho não pode aparecer duas vezes na mesma reserva")
		}

		referenciasVistas[trecho] = true
	}

	return &Reserva{
		ID:                 id,
		PassageiroID:       passageiroID,
		QuantidadeAssentos: quantidadeAssentos,
		Status:             StatusReservaConfirmada,
		trechos:            append([]ReferenciaTrecho(nil), trechos...),
	}, nil
}

func (reserva Reserva) Trechos() []ReferenciaTrecho {
	return append([]ReferenciaTrecho(nil), reserva.trechos...)
}

func (reserva Reserva) EstaConfirmada() bool {
	return reserva.Status == StatusReservaConfirmada
}

func (reserva *Reserva) Cancelar() error {
	if !reserva.EstaConfirmada() {
		return errors.New("a reserva não está confirmada")
	}

	reserva.Status = StatusReservaCancelada
	return nil
}
