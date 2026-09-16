package store

import (
	"errors"
	"sync"
	"vaijunto/internal/domain"
)

type CatalogoCaronas struct {
	mu      sync.Mutex
	caronas map[string]*domain.Carona
}

func NovoCatalogoCaronas() *CatalogoCaronas {
	return &CatalogoCaronas{
		caronas: make(map[string]*domain.Carona),
	}
}

func (catalogo *CatalogoCaronas) Adicionar(carona *domain.Carona) error {
	if carona == nil {
		return errors.New("a carona nao pode ser nula")
	}

	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	_, existe := catalogo.caronas[carona.ID]

	if existe {
		return errors.New("já existe uma carona com esse id")
	}

	catalogo.caronas[carona.ID] = carona

	return nil
}

func (catalogo *CatalogoCaronas) BuscarPorID(id string) (*domain.Carona, bool) {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	carona, existe := catalogo.caronas[id]

	if !existe {
		return nil, false
	}
	return carona, true
}

func (catalogo *CatalogoCaronas) ListarTrechos() []domain.Trecho {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	trechos := make([]domain.Trecho, 0)

	for _, carona := range catalogo.caronas {
		trechos = append(trechos, carona.Trechos()...)
	}

	return trechos
}
