package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"time"
	"vaijunto/internal/domain"
	"vaijunto/internal/protocol"
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

		if requisicao.Operacao == "ping" {
			resposta := protocol.Resposta{Mensagem: "pong"}
			json.NewEncoder(conexao).Encode(resposta)
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
