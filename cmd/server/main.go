package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"time"
	"vaijunto/internal/domain"
	"vaijunto/internal/protocol"
	"vaijunto/internal/search"
	"vaijunto/internal/store"
)

const enderecoServidor = ":8080"

func main() {
	listener, err := net.Listen("tcp", enderecoServidor)
	if err != nil {
		log.Fatalf("nao foi possivel iniciar o servidor: %v", err)
	}
	defer listener.Close()

	catalogo := store.NovoCatalogoCaronas()

	log.Printf("servidor ouvindo em %s", enderecoServidor)

	for {
		conexao, err := listener.Accept()
		if err != nil {
			log.Printf("erro ao aceitar conexao: %v", err)
			continue
		}

		go tratarConexao(conexao, catalogo)
	}
}

func tratarConexao(conexao net.Conn, catalogo *store.CatalogoCaronas) {
	defer conexao.Close()

	leitor := bufio.NewScanner(conexao)

	for leitor.Scan() {
		var requisicao protocol.Requisicao
		err := json.Unmarshal(leitor.Bytes(), &requisicao)
		if err != nil {
			continue
		}

		var resposta protocol.Resposta
		if requisicao.Operacao == "ping" {
			resposta = protocol.Resposta{
				Sucesso:  true,
				Mensagem: "pong",
			}
		} else if requisicao.Operacao == "criar_carona" {
			resposta = tratarCriarCarona(requisicao, catalogo)
		} else if requisicao.Operacao == "buscar_itinerarios" {
			resposta = tratarBuscarItinerarios(requisicao, catalogo)
		} else {
			resposta = protocol.Resposta{
				Sucesso:  false,
				Mensagem: "operação desconhecida",
			}
		}

		err = json.NewEncoder(conexao).Encode(resposta)

		if err != nil {
			return
		}
	}
}

func tratarCriarCarona(requisicao protocol.Requisicao, catalogo *store.CatalogoCaronas) protocol.Resposta {
	var dados protocol.CriarCarona
	err := json.Unmarshal(requisicao.Dados, &dados)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: "dados da carona em JSON inválidos",
		}
	}

	horarioSaida, err := time.Parse(time.RFC3339, dados.HorarioSaida)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: "horário de saída inválido",
		}
	}

	carona, err := domain.NovaCarona(
		dados.ID,
		dados.MotoristaID,
		horarioSaida,
		dados.Rota,
		dados.Capacidade,
		dados.PrecosCentavos,
	)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: err.Error(),
		}
	}

	err = catalogo.Adicionar(carona)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: err.Error(),
		}
	}

	return protocol.Resposta{
		Sucesso:  true,
		Mensagem: "carona criada com sucesso",
	}
}

func tratarBuscarItinerarios(requisicao protocol.Requisicao, catalogo *store.CatalogoCaronas) protocol.Resposta {
	var dados protocol.BuscarItinerarios
	err := json.Unmarshal(requisicao.Dados, &dados)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: "dados da busca em JSON inválidos",
		}
	}

	trechos := catalogo.ListarTrechos()
	itinerarios, err := search.BuscarItinerarios(
		trechos,
		dados.Origem,
		dados.Destino,
		dados.QuantidadeAssentos,
	)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: err.Error(),
		}
	}

	itinerariosEncontrados := make([]protocol.ItinerarioEncontrado, 0, len(itinerarios))

	for _, itinerario := range itinerarios {
		trechosEncontrados := make([]protocol.TrechoEncontrado, 0, len(itinerario.Trechos))

		for _, trecho := range itinerario.Trechos {
			trechosEncontrados = append(trechosEncontrados, protocol.TrechoEncontrado{
				CaronaID:            trecho.CaronaID,
				Ordem:               trecho.Ordem,
				Origem:              trecho.Origem,
				Destino:             trecho.Destino,
				PrecoCentavos:       trecho.PrecoCentavos,
				AssentosDisponiveis: trecho.AssentosDisponiveis,
			})
		}

		itinerariosEncontrados = append(itinerariosEncontrados, protocol.ItinerarioEncontrado{
			Origem:             itinerario.Origem,
			Destino:            itinerario.Destino,
			Trechos:            trechosEncontrados,
			PrecoTotalCentavos: itinerario.PrecoTotalCentavos,
		})
	}

	dadosJSON, err := json.Marshal(itinerariosEncontrados)
	if err != nil {
		return protocol.Resposta{
			Sucesso:  false,
			Mensagem: "não foi possível preparar a resposta da busca",
		}
	}

	mensagem := "itinerários encontrados"
	if len(itinerariosEncontrados) == 0 {
		mensagem = "nenhum itinerário encontrado"
	}

	return protocol.Resposta{
		Sucesso:  true,
		Mensagem: mensagem,
		Dados:    dadosJSON,
	}
}
