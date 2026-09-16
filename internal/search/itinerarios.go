package search

import (
	"errors"
	"sort"
	"strings"

	"vaijunto/internal/domain"
)

const (
	// Limites evitam que uma busca em um grafo muito conectado cresça sem
	// controle e consuma recursos do servidor.
	maximoTrechosPorItinerario = 8
	maximoCandidatos           = 1000
	maximoResultados           = 20
)

func BuscarItinerarios(trechos []domain.Trecho, origem string, destino string, quantidadeAssentos int) ([]domain.Itinerario, error) {
	// Cada cidade é um vértice e cada trecho disponível é uma aresta dirigida.
	origem = normalizarCidade(origem)
	destino = normalizarCidade(destino)

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

	// A interface apresenta primeiro o itinerário de menor preço total.
	sort.Slice(itinerarios, func(i int, j int) bool {
		return itinerarios[i].PrecoTotalCentavos < itinerarios[j].PrecoTotalCentavos
	})

	if len(itinerarios) > maximoResultados {
		itinerarios = itinerarios[:maximoResultados]
	}

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
	// Esta é uma busca em profundidade (DFS): o caminho atual é estendido até
	// chegar ao destino ou até uma condição de parada.
	if len(*itinerarios) >= maximoCandidatos {
		return
	}

	if cidadeAtual == destino {
		// Copia o slice porque a recursão continuará montando outros caminhos.
		itinerario := domain.Itinerario{
			Origem:  caminhoAtual[0].Origem,
			Destino: caminhoAtual[len(caminhoAtual)-1].Destino,
			Trechos: append([]domain.Trecho(nil), caminhoAtual...),
		}

		itinerario.PrecoTotalCentavos = itinerario.CalcularPrecoTotal()

		*itinerarios = append(*itinerarios, itinerario)
		return
	}

	if len(caminhoAtual) >= maximoTrechosPorItinerario {
		return
	}

	for _, trecho := range trechosDisponiveis {
		origemTrecho := normalizarCidade(trecho.Origem)
		if origemTrecho != cidadeAtual {
			continue
		}

		if !trecho.TemVagas(quantidadeAssentos) {
			continue
		}

		destinoTrecho := normalizarCidade(trecho.Destino)

		// Não visita a mesma cidade duas vezes no mesmo caminho, evitando ciclos.
		if cidadesVisitadas[destinoTrecho] {
			continue
		}

		// Marca antes da chamada recursiva e desmarca ao voltar: é o backtracking.
		cidadesVisitadas[destinoTrecho] = true

		proximoCaminho := append(caminhoAtual, trecho)

		buscarCaminhos(
			trechosDisponiveis,
			destinoTrecho,
			destino,
			quantidadeAssentos,
			proximoCaminho,
			cidadesVisitadas,
			itinerarios,
		)

		delete(cidadesVisitadas, destinoTrecho)
	}
}

func normalizarCidade(cidade string) string {
	return strings.ToLower(strings.TrimSpace(cidade))
}
