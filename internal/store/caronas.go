package store

import (
	"errors"
	"sort"
	"sync"

	"vaijunto/internal/domain"
)

type PassageiroNoTrecho struct {
	PassageiroID       string
	QuantidadeAssentos int
}

type SituacaoTrecho struct {
	Trecho      domain.Trecho
	Passageiros []PassageiroNoTrecho
}

type SituacaoCarona struct {
	ID      string
	Trechos []SituacaoTrecho
}

type CatalogoCaronas struct {
	mu       sync.Mutex
	caronas  map[string]*domain.Carona
	reservas map[string]*domain.Reserva
}

func NovoCatalogoCaronas() *CatalogoCaronas {
	return &CatalogoCaronas{
		caronas:  make(map[string]*domain.Carona),
		reservas: make(map[string]*domain.Reserva),
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

func (catalogo *CatalogoCaronas) ConfirmarReserva(
	idReserva string,
	passageiroID string,
	quantidadeAssentos int,
	referencias []domain.ReferenciaTrecho,
) (*domain.Reserva, error) {
	reserva, err := domain.NovaReserva(
		idReserva,
		passageiroID,
		quantidadeAssentos,
		referencias,
	)
	if err != nil {
		return nil, err
	}

	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	_, existe := catalogo.reservas[idReserva]
	if existe {
		return nil, errors.New("já existe uma reserva com esse ID")
	}

	trechosReserva := reserva.Trechos()

	for _, referencia := range trechosReserva {
		carona, encontrada := catalogo.caronas[referencia.CaronaID]
		if !encontrada {
			return nil, errors.New("carona da reserva não encontrada")
		}

		if !carona.TemVagasNoTrecho(referencia.Ordem, quantidadeAssentos) {
			return nil, errors.New("não há vagas suficientes em todos os trechos")
		}
	}

	referenciasReservadas := make([]domain.ReferenciaTrecho, 0, len(trechosReserva))

	for _, referencia := range trechosReserva {
		carona := catalogo.caronas[referencia.CaronaID]

		err = carona.ReservarNoTrecho(referencia.Ordem, quantidadeAssentos)
		if err != nil {
			for _, referenciaReservada := range referenciasReservadas {
				caronaReservada := catalogo.caronas[referenciaReservada.CaronaID]
				caronaReservada.CancelarNoTrecho(referenciaReservada.Ordem, quantidadeAssentos)
			}

			return nil, errors.New("não foi possível confirmar a reserva")
		}

		referenciasReservadas = append(referenciasReservadas, referencia)
	}

	catalogo.reservas[reserva.ID] = reserva

	return reserva, nil
}

func (catalogo *CatalogoCaronas) BuscarReservaPorID(id string) (*domain.Reserva, bool) {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	reserva, existe := catalogo.reservas[id]
	if !existe {
		return nil, false
	}

	return reserva, true
}

func (catalogo *CatalogoCaronas) ListarReservasDoPassageiro(passageiroID string) []domain.Reserva {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	reservas := make([]domain.Reserva, 0)

	for _, reserva := range catalogo.reservas {
		if reserva.PassageiroID == passageiroID {
			reservas = append(reservas, *reserva)
		}
	}

	sort.Slice(reservas, func(i int, j int) bool {
		return reservas[i].ID < reservas[j].ID
	})

	return reservas
}

func (catalogo *CatalogoCaronas) ListarCaronasDoMotorista(motoristaID string) []SituacaoCarona {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	caronas := make([]SituacaoCarona, 0)

	for _, carona := range catalogo.caronas {
		if carona.MotoristaID() != motoristaID {
			continue
		}

		situacao := SituacaoCarona{
			ID:      carona.ID,
			Trechos: make([]SituacaoTrecho, 0),
		}

		for _, trecho := range carona.Trechos() {
			situacaoTrecho := SituacaoTrecho{
				Trecho:      trecho,
				Passageiros: make([]PassageiroNoTrecho, 0),
			}

			for _, reserva := range catalogo.reservas {
				if !reserva.EstaConfirmada() {
					continue
				}

				for _, referencia := range reserva.Trechos() {
					if referencia.CaronaID == carona.ID && referencia.Ordem == trecho.Ordem {
						situacaoTrecho.Passageiros = append(
							situacaoTrecho.Passageiros,
							PassageiroNoTrecho{
								PassageiroID:       reserva.PassageiroID,
								QuantidadeAssentos: reserva.QuantidadeAssentos,
							},
						)
					}
				}
			}

			sort.Slice(situacaoTrecho.Passageiros, func(i int, j int) bool {
				return situacaoTrecho.Passageiros[i].PassageiroID < situacaoTrecho.Passageiros[j].PassageiroID
			})

			situacao.Trechos = append(situacao.Trechos, situacaoTrecho)
		}

		caronas = append(caronas, situacao)
	}

	sort.Slice(caronas, func(i int, j int) bool {
		return caronas[i].ID < caronas[j].ID
	})

	return caronas
}

func (catalogo *CatalogoCaronas) CancelarReserva(idReserva string, passageiroID string) error {
	catalogo.mu.Lock()
	defer catalogo.mu.Unlock()

	reserva, existe := catalogo.reservas[idReserva]
	if !existe {
		return errors.New("reserva não encontrada")
	}

	if reserva.PassageiroID != passageiroID {
		return errors.New("a reserva não pertence ao passageiro informado")
	}

	if !reserva.EstaConfirmada() {
		return errors.New("a reserva não está confirmada")
	}

	referenciasCanceladas := make([]domain.ReferenciaTrecho, 0, len(reserva.Trechos()))

	for _, referencia := range reserva.Trechos() {
		carona, encontrada := catalogo.caronas[referencia.CaronaID]
		if !encontrada {
			catalogo.desfazerCancelamento(referenciasCanceladas, reserva.QuantidadeAssentos)
			return errors.New("carona da reserva não encontrada")
		}

		err := carona.CancelarNoTrecho(referencia.Ordem, reserva.QuantidadeAssentos)
		if err != nil {
			catalogo.desfazerCancelamento(referenciasCanceladas, reserva.QuantidadeAssentos)
			return errors.New("não foi possível cancelar a reserva")
		}

		referenciasCanceladas = append(referenciasCanceladas, referencia)
	}

	err := reserva.Cancelar()
	if err != nil {
		catalogo.desfazerCancelamento(referenciasCanceladas, reserva.QuantidadeAssentos)
		return errors.New("não foi possível cancelar a reserva")
	}

	return nil
}

func (catalogo *CatalogoCaronas) desfazerCancelamento(
	referencias []domain.ReferenciaTrecho,
	quantidadeAssentos int,
) {
	for _, referencia := range referencias {
		carona := catalogo.caronas[referencia.CaronaID]
		carona.ReservarNoTrecho(referencia.Ordem, quantidadeAssentos)
	}
}
