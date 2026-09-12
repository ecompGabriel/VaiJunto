package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"vaijunto/internal/protocol"
)

func main() {
	conexao, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("nao foi possivel conectar ao servidor: %v", err)
	}
	defer conexao.Close()

	requisicao := protocol.Requisicao{
		Operacao: "ping",
	}

	if err := json.NewEncoder(conexao).Encode(requisicao); err != nil {
		log.Fatalf("nao foi possivel enviar requisicao: %v", err)
	}

	var resposta protocol.Resposta
	if err := json.NewDecoder(conexao).Decode(&resposta); err != nil {
		log.Fatalf("nao foi possivel ler resposta: %v", err)
	}

	fmt.Println("Resposta do servidor:", resposta.Mensagem)
}
