package domain

import "errors"

type Trecho struct {
	CaronaID            string
	Ordem               int
	Origem              string
	Destino             string
	Capacidade          int
	AssentosDisponiveis int
	PrecoCentavos       int64
}

func (t Trecho) TemVagas(quantidadeSolicitada int) bool {

	if quantidadeSolicitada <= 0 {
		return false
	}

	return t.AssentosDisponiveis >= quantidadeSolicitada
}

func (t *Trecho) Reservar(quantidadeSolicitada int) error {
	if quantidadeSolicitada <= 0 {
		return errors.New("Pedido inválido")
	}
	if quantidadeSolicitada > t.AssentosDisponiveis {
		return errors.New("Quantidade solicitada maior que a quantidade de assentos disponíveis")
	}
	t.AssentosDisponiveis -= quantidadeSolicitada

	return nil
}

func (t *Trecho) CancelarReserva(quantidadeCancelada int) error {
	if quantidadeCancelada <= 0 {
		return errors.New("Cancelamento inválido")
	}
	if t.AssentosDisponiveis+quantidadeCancelada > t.Capacidade {
		return errors.New("Cancelamento ultrapassa a capacidade do trecho")
	}
	t.AssentosDisponiveis += quantidadeCancelada
	return nil
}
