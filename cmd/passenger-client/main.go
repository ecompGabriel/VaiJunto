package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"vaijunto/internal/protocol"
)

func main() {
	conexao, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("nao foi possivel conectar ao servidor: %v", err)
	}
	defer conexao.Close()

	leitor := bufio.NewReader(os.Stdin)
	origem := lerTexto(leitor, "Cidade de origem: ")
	destino := lerTexto(leitor, "Cidade de destino: ")
	textoQuantidade := lerTexto(leitor, "Quantidade de assentos: ")

	quantidadeAssentos, err := strconv.Atoi(textoQuantidade)
	if err != nil || quantidadeAssentos <= 0 {
		log.Fatalf("a quantidade de assentos deve ser um número inteiro positivo")
	}

	dados := protocol.BuscarItinerarios{
		Origem:             origem,
		Destino:            destino,
		QuantidadeAssentos: quantidadeAssentos,
	}

	dadosJSON, err := json.Marshal(dados)
	if err != nil {
		log.Fatalf("não foi possível transformar a busca em JSON: %v", err)
	}

	requisicao := protocol.Requisicao{
		Operacao: "buscar_itinerarios",
		Dados:    dadosJSON,
	}

	if err := json.NewEncoder(conexao).Encode(requisicao); err != nil {
		log.Fatalf("nao foi possivel enviar requisicao: %v", err)
	}

	var resposta protocol.Resposta
	if err := json.NewDecoder(conexao).Decode(&resposta); err != nil {
		log.Fatalf("nao foi possivel ler resposta: %v", err)
	}

	fmt.Println("Resposta do servidor:", resposta.Mensagem)

	if !resposta.Sucesso {
		log.Fatalf("o servidor recusou a busca: %s", resposta.Mensagem)
	}

	var itinerarios []protocol.ItinerarioEncontrado

	err = json.Unmarshal(resposta.Dados, &itinerarios)
	if err != nil {
		log.Fatalf("não foi possível decodificar os itinerários recebidos: %v", err)
	}

	if len(itinerarios) == 0 {
		fmt.Println("Nenhum itinerário encontrado.")
		return
	}

	for indice, itinerario := range itinerarios {
		fmt.Printf("\nOpção %d\n", indice+1)
		fmt.Printf("De %s até %s\n", itinerario.Origem, itinerario.Destino)
		fmt.Printf(
			"Preço total: R$ %d,%02d\n",
			itinerario.PrecoTotalCentavos/100,
			itinerario.PrecoTotalCentavos%100,
		)

		fmt.Println("Trechos:")

		for _, trecho := range itinerario.Trechos {
			fmt.Printf(
				"- %s → %s — R$ %d,%02d — Assentos disponíveis: %d\n",
				trecho.Origem,
				trecho.Destino,
				trecho.PrecoCentavos/100,
				trecho.PrecoCentavos%100,
				trecho.AssentosDisponiveis,
			)
		}
	}
}

func lerTexto(leitor *bufio.Reader, pergunta string) string {
	for {
		fmt.Print(pergunta)

		texto, err := leitor.ReadString('\n')
		if err != nil {
			log.Fatalf("não foi possível ler a entrada: %v", err)
		}

		texto = strings.TrimSpace(texto)
		if texto == "" {
			fmt.Println("Este campo não pode ficar vazio")
			continue
		}

		return texto
	}
}
