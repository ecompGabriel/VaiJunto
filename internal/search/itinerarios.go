package search

import (
	"errors"
	"sort"
	"strings"

	"vaijunto/internal/domain"
)

func BuscarItinerarios(trechos []domain.Trecho, origem string, destino string, quantidadeAssentos int) ([]domain.Itinerario, error) {
	origem = strings.TrimSpace(origem)
	destino = strings.TrimSpace(destino)

	if origem == "" || destino == "" {
		return nil, errors.New("origem e destino são obrigatórios")
	}

	if origem == destino {
		return nil, errors.New("origem e destino devem ser diferentes")
	}

	if quantidadeAssentos <= 0 {
		return nil, errors.New("a quantidade de assentos deve ser positiva")
	}

	itinerarios := make([]domain.Itinerario, 0)

	cidadesVisitadas := make(map[string]bool)
	cidadesVisitadas[origem] = true

	buscarCaminhos(
		trechos,
		origem,
		destino,
		quantidadeAssentos,
		nil,
		cidadesVisitadas,
		&itinerarios,
	)

	sort.Slice(itinerarios, func(i int, j int) bool {
		return itinerarios[i].PrecoTotalCentavos < itinerarios[j].PrecoTotalCentavos
	})

	return itinerarios, nil
}

func buscarCaminhos(
	trechosDisponiveis []domain.Trecho,
	cidadeAtual string,
	destino string,
	quantidadeAssentos int,
	caminhoAtual []domain.Trecho,
	cidadesVisitadas map[string]bool,
	itinerarios *[]domain.Itinerario,
) {
	if cidadeAtual == destino {
		itinerario := domain.Itinerario{
			Origem:  caminhoAtual[0].Origem,
			Destino: destino,
			Trechos: append([]domain.Trecho(nil), caminhoAtual...),
		}

		itinerario.PrecoTotalCentavos = itinerario.CalcularPrecoTotal()

		*itinerarios = append(*itinerarios, itinerario)
		return
	}

	for _, trecho := range trechosDisponiveis {
		if trecho.Origem != cidadeAtual {
			continue
		}

		if !trecho.TemVagas(quantidadeAssentos) {
			continue
		}

		if cidadesVisitadas[trecho.Destino] {
			continue
		}

		cidadesVisitadas[trecho.Destino] = true

		proximoCaminho := append(caminhoAtual, trecho)

		buscarCaminhos(
			trechosDisponiveis,
			trecho.Destino,
			destino,
			quantidadeAssentos,
			proximoCaminho,
			cidadesVisitadas,
			itinerarios,
		)

		delete(cidadesVisitadas, trecho.Destino)
	}
}
