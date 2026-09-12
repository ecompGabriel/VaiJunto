package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"

	"vaijunto/internal/protocol"
)

const enderecoServidor = ":8080"

func main() {
	listener, err := net.Listen("tcp", enderecoServidor)
	if err != nil {
		log.Fatalf("nao foi possivel iniciar o servidor: %v", err)
	}
	defer listener.Close()

	log.Printf("servidor ouvindo em %s", enderecoServidor)

	for {
		conexao, err := listener.Accept()
		if err != nil {
			log.Printf("erro ao aceitar conexao: %v", err)
			continue
		}

		go tratarConexao(conexao)
	}
}

func tratarConexao(conexao net.Conn) {
	defer conexao.Close()

	leitor := bufio.NewScanner(conexao)

	for leitor.Scan() {
		var requisicao protocol.Requisicao
		if err := json.Unmarshal(leitor.Bytes(), &requisicao); err != nil {
			continue
		}

		if requisicao.Operacao == "ping" {
			resposta := protocol.Resposta{Mensagem: "pong"}
			json.NewEncoder(conexao).Encode(resposta)
		}
	}
}
