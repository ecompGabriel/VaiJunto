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

func (carona Carona) Trechos() []Trecho {
	return append([]Trecho(nil), carona.trechos...)
}

func (carona Carona) MotoristaID() string {
	return carona.motoristaID
}

func (carona Carona) TemVagasNoTrecho(ordem int, quantidade int) bool {
	for _, trecho := range carona.trechos {
		if trecho.Ordem == ordem {
			return trecho.TemVagas(quantidade)
		}
	}

	return false
}

func (carona *Carona) ReservarNoTrecho(ordem int, quantidade int) error {
	trecho, err := carona.encontrarTrechoPorOrdem(ordem)
	if err != nil {
		return err
	}

	return trecho.Reservar(quantidade)
}

func (carona *Carona) CancelarNoTrecho(ordem int, quantidade int) error {
	trecho, err := carona.encontrarTrechoPorOrdem(ordem)
	if err != nil {
		return err
	}

	return trecho.CancelarReserva(quantidade)
}

func (carona *Carona) encontrarTrechoPorOrdem(ordem int) (*Trecho, error) {
	for indice := range carona.trechos {
		if carona.trechos[indice].Ordem == ordem {
			return &carona.trechos[indice], nil
		}
	}

	return nil, errors.New("trecho não encontrado na carona")
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

func NovaCarona(id string, motoristaID string, horarioSaida time.Time, rota []string, capacidade int, precosCentavos []int64) (*Carona, error) {
	if len(rota) < 2 {
		return nil, errors.New("a rota precisa ter ao menos duas cidades")
	}

	if capacidade <= 0 {
		return nil, errors.New("a capacidade precisa ser maior que zero")
	}

	quantidadeTrechos := len(rota) - 1
	if len(precosCentavos) != quantidadeTrechos {
		return nil, errors.New("a quantidade de preços deve ser igual à quantidade de trechos")
	}

	trechos := make([]Trecho, 0, quantidadeTrechos)
	for i := 0; i < quantidadeTrechos; i++ {
		if precosCentavos[i] < 0 {
			return nil, errors.New("o preço de um trecho não pode ser negativo")
		}

		trechos = append(trechos, Trecho{
			CaronaID:            id,
			Ordem:               i,
			Origem:              rota[i],
			Destino:             rota[i+1],
			Capacidade:          capacidade,
			AssentosDisponiveis: capacidade,
			PrecoCentavos:       precosCentavos[i],
		})
	}

	return &Carona{
		ID:           id,
		motoristaID:  motoristaID,
		horarioSaida: horarioSaida,
		rota:         append([]string(nil), rota...),
		trechos:      trechos,
	}, nil
}
