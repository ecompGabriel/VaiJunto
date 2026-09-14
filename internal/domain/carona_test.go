package domain

import "testing"

func caronaDeTeste() Carona {
	return Carona{
		ID:   "carona-1",
		rota: []string{"Feira", "Alagoinhas", "Salvador"},
		trechos: []Trecho{
			{
				CaronaID:            "carona-1",
				Ordem:               0,
				Origem:              "Feira",
				Destino:             "Alagoinhas",
				Capacidade:          4,
				AssentosDisponiveis: 4,
				PrecoCentavos:       1500,
			},
			{
				CaronaID:            "carona-1",
				Ordem:               1,
				Origem:              "Alagoinhas",
				Destino:             "Salvador",
				Capacidade:          4,
				AssentosDisponiveis: 4,
				PrecoCentavos:       2000,
			},
		},
	}
}

func TestCaronaTemCidade(t *testing.T) {
	carona := caronaDeTeste()

	if !carona.TemCidade("Alagoinhas") {
		t.Error("esperava encontrar Alagoinhas na rota")
	}

	if carona.TemCidade("Camaçari") {
		t.Error("não esperava encontrar Camaçari na rota")
	}
}

func TestCaronaPosicaoCidade(t *testing.T) {
	carona := caronaDeTeste()

	posicao, encontrou := carona.PosicaoCidade("Salvador")
	if !encontrou {
		t.Fatal("esperava encontrar Salvador na rota")
	}

	if posicao != 2 {
		t.Errorf("posição de Salvador = %d; esperava 2", posicao)
	}

	_, encontrou = carona.PosicaoCidade("Camaçari")
	if encontrou {
		t.Error("não esperava encontrar Camaçari na rota")
	}
}

func TestCaronaTrechosEntre(t *testing.T) {
	carona := caronaDeTeste()

	trechos, err := carona.TrechosEntre("Feira", "Salvador")
	if err != nil {
		t.Fatalf("não esperava erro ao buscar trechos: %v", err)
	}

	if len(trechos) != 2 {
		t.Fatalf("quantidade de trechos = %d; esperava 2", len(trechos))
	}

	if trechos[0].Origem != "Feira" || trechos[0].Destino != "Alagoinhas" {
		t.Error("primeiro trecho não é Feira para Alagoinhas")
	}

	if trechos[1].Origem != "Alagoinhas" || trechos[1].Destino != "Salvador" {
		t.Error("segundo trecho não é Alagoinhas para Salvador")
	}
}

func TestCaronaTrechosEntreRejeitaCidadesInvalidas(t *testing.T) {
	carona := caronaDeTeste()

	_, err := carona.TrechosEntre("Feira", "Camaçari")
	if err == nil {
		t.Error("esperava erro para destino fora da rota")
	}

	_, err = carona.TrechosEntre("Salvador", "Feira")
	if err == nil {
		t.Error("esperava erro quando a origem vem depois do destino")
	}
}

func TestCaronaPrecosEntre(t *testing.T) {
	carona := caronaDeTeste()

	preco, err := carona.PrecosEntre("Feira", "Salvador")
	if err != nil {
		t.Fatalf("não esperava erro ao calcular preço: %v", err)
	}

	if preco != 3500 {
		t.Errorf("preço = %d centavos; esperava 3500", preco)
	}
}
