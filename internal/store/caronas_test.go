package store

import (
	"testing"
	"time"

	"vaijunto/internal/domain"
)

func novaCaronaDeTeste(t *testing.T, id string) *domain.Carona {
	t.Helper()

	carona, err := domain.NovaCarona(
		id,
		"motorista-1",
		time.Date(2026, time.September, 17, 8, 0, 0, 0, time.UTC),
		[]string{"Feira", "Salvador"},
		4,
		[]int64{2000},
	)
	if err != nil {
		t.Fatalf("não esperava erro ao criar carona de teste: %v", err)
	}

	return carona
}

func TestCatalogoAdicionaCarona(t *testing.T) {
	catalogo := NovoCatalogoCaronas()
	carona := novaCaronaDeTeste(t, "carona-1")

	err := catalogo.Adicionar(carona)
	if err != nil {
		t.Fatalf("não esperava erro ao adicionar carona: %v", err)
	}

	caronaEncontrada, encontrada := catalogo.BuscarPorID("carona-1")
	if !encontrada {
		t.Fatal("esperava encontrar a carona adicionada")
	}

	if caronaEncontrada != carona {
		t.Error("a carona encontrada não é a mesma que foi adicionada")
	}
}

func TestCatalogoRejeitaIDDuplicado(t *testing.T) {
	catalogo := NovoCatalogoCaronas()

	err := catalogo.Adicionar(novaCaronaDeTeste(t, "carona-1"))
	if err != nil {
		t.Fatalf("não esperava erro ao adicionar a primeira carona: %v", err)
	}

	err2 := catalogo.Adicionar(novaCaronaDeTeste(t, "carona-1"))
	if err2 == nil {
		t.Error("esperava erro ao adicionar uma carona com ID duplicado")
	}
}

func TestCatalogoBuscaCaronaPorIDInexistente(t *testing.T) {
	catalogo := NovoCatalogoCaronas()

	carona, encontrada := catalogo.BuscarPorID("carona-inexistente")

	if encontrada {
		t.Error("não esperava encontrar uma carona inexistente")
	}

	if carona != nil {
		t.Error("esperava nil para uma carona inexistente")
	}
}
