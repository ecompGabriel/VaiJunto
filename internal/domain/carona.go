package domain

import (
	"errors"
	"time"
)

type Carona struct {
	ID           string
	motoristaID  string
	horarioSaida time.Time
	rota         []string
	trechos      []Trecho
}

func (carona Carona) TemCidade(cidade string) bool {

	for _, c := range carona.rota {
		if c == cidade {
			return true
		}
	}
	return false
}

func (carona Carona) PosicaoCidade(cidade string) (int, bool) {

	for i, c := range carona.rota {
		if c == cidade {
			return i, true
		}
	}
	return 0, false
}

func (carona Carona) TrechosEntre(origem string, destino string) ([]Trecho, error) {
	posicaoOrigem, encontrouOrigem := carona.PosicaoCidade(origem)
	posicaoDestino, encontrouDestino := carona.PosicaoCidade(destino)

	if !encontrouOrigem || !encontrouDestino {
		return nil, errors.New("origem ou destino não pertencem à rota")
	}

	if posicaoOrigem >= posicaoDestino {
		return nil, errors.New("a origem deve estar antes do destino na rota")
	}

	return carona.trechos[posicaoOrigem:posicaoDestino], nil
}

func (carona Carona) PrecosEntre(origem string, destino string) (int64, error) {
	trechos, err := carona.TrechosEntre(origem, destino)
	if err != nil {
		return 0, err
	}

	var precoTotal int64
	for _, trecho := range trechos {
		precoTotal += trecho.PrecoCentavos
	}
	return precoTotal, nil
}
