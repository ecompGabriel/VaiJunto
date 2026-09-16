package search

import (
	"testing"

	"vaijunto/internal/domain"
)

func TestBuscarItinerariosEncontraCaminhoComDoisTrechos(t *testing.T) {
	trechos := []domain.Trecho{
		{
			CaronaID:            "carona-1",
			Ordem:               1,
			Origem:              "Feira",
			Destino:             "Alagoinhas",
			Capacidade:          3,
			AssentosDisponiveis: 3,
			PrecoCentavos:       1000,
		},
		{
			CaronaID:            "carona-1",
			Ordem:               2,
			Origem:              "Alagoinhas",
			Destino:             "Salvador",
			Capacidade:          3,
			AssentosDisponiveis: 3,
			PrecoCentavos:       1500,
		},
	}

	itinerarios, err := BuscarItinerarios(trechos, "Feira", "Salvador", 1)
	if err != nil {
		t.Fatalf("não esperava erro na busca: %v", err)
	}

	if len(itinerarios) != 1 {
		t.Fatalf("quantidade de itinerários = %d; esperava 1", len(itinerarios))
	}

	itinerario := itinerarios[0]

	if len(itinerario.Trechos) != 2 {
		t.Errorf("quantidade de trechos = %d; esperava 2", len(itinerario.Trechos))
	}

	if itinerario.PrecoTotalCentavos != 2500 {
		t.Errorf(
			"preço total = %d centavos; esperava 2500",
			itinerario.PrecoTotalCentavos,
		)
	}
}

func TestBuscarItinerariosOrdenaPeloMenorPreco(t *testing.T) {
	trechos := []domain.Trecho{
		{
			CaronaID:            "carona-cara",
			Origem:              "Feira",
			Destino:             "Salvador",
			Capacidade:          3,
			AssentosDisponiveis: 3,
			PrecoCentavos:       3000,
		},
		{
			CaronaID:            "carona-barata",
			Origem:              "Feira",
			Destino:             "Salvador",
			Capacidade:          3,
			AssentosDisponiveis: 3,
			PrecoCentavos:       2000,
		},
	}

	itinerarios, err := BuscarItinerarios(trechos, "Feira", "Salvador", 1)
	if err != nil {
		t.Fatalf("não esperava erro na busca: %v", err)
	}

	if len(itinerarios) != 2 {
		t.Fatalf("quantidade de itinerários = %d; esperava 2", len(itinerarios))
	}

	if itinerarios[0].PrecoTotalCentavos != 2000 {
		t.Errorf(
			"primeiro itinerário custa %d centavos; esperava 2000",
			itinerarios[0].PrecoTotalCentavos,
		)
	}
}

func TestBuscarItinerariosIgnoraTrechoSemVagas(t *testing.T) {
	trechos := []domain.Trecho{
		{
			CaronaID:            "carona-1",
			Origem:              "Feira",
			Destino:             "Salvador",
			Capacidade:          2,
			AssentosDisponiveis: 0,
			PrecoCentavos:       2000,
		},
	}

	itinerarios, err := BuscarItinerarios(trechos, "Feira", "Salvador", 2)
	if err != nil {
		t.Fatalf("não esperava erro na busca: %v", err)
	}

	if len(itinerarios) != 0 {
		t.Errorf("quantidade de itinerários = %d; esperava 0", len(itinerarios))
	}
}

func TestBuscarItinerariosNaoDiferenciaMaiusculasEMinusculas(t *testing.T) {
	trechos := []domain.Trecho{
		{
			CaronaID:            "carona-1",
			Origem:              "Feira",
			Destino:             "Salvador",
			Capacidade:          3,
			AssentosDisponiveis: 3,
			PrecoCentavos:       2000,
		},
	}

	itinerarios, err := BuscarItinerarios(trechos, "  FEIRA ", "sAlVaDoR", 1)
	if err != nil {
		t.Fatalf("não esperava erro na busca: %v", err)
	}

	if len(itinerarios) != 1 {
		t.Fatalf("quantidade de itinerários = %d; esperava 1", len(itinerarios))
	}

	if itinerarios[0].Origem != "Feira" || itinerarios[0].Destino != "Salvador" {
		t.Errorf(
			"itinerário retornado = %s até %s; esperava Feira até Salvador",
			itinerarios[0].Origem,
			itinerarios[0].Destino,
		)
	}
}

func TestBuscarItinerariosRejeitaQuantidadeInvalida(t *testing.T) {
	itinerarios, err := BuscarItinerarios(nil, "Feira", "Salvador", 0)

	if err == nil {
		t.Fatal("esperava erro para quantidade de assentos igual a zero")
	}

	if itinerarios != nil {
		t.Errorf("itinerários = %v; esperava nil", itinerarios)
	}
}
