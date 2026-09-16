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
	"time"

	"vaijunto/internal/protocol"
)

func main() {
	leitor := bufio.NewReader(os.Stdin)

	id := lerTexto(leitor, "ID da carona: ")
	motoristaID := lerTexto(leitor, "ID do motorista: ")
	horarioSaida := lerHorarioSaida(leitor)
	rota := lerRota(leitor)
	capacidade := lerCapacidade(leitor)
	precosCentavos := lerPrecosCentavos(leitor, rota)

	conexao, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("nao foi possivel conectar ao servidor: %v", err)
	}
	defer conexao.Close()

	dados := protocol.CriarCarona{
		ID:             id,
		MotoristaID:    motoristaID,
		HorarioSaida:   horarioSaida,
		Rota:           rota,
		Capacidade:     capacidade,
		PrecosCentavos: precosCentavos,
	}

	dadosJSON, err := json.Marshal(dados)

	if err != nil {
		log.Fatalf("não foi possível transformar os dados da carona em JSON: %v", err)
	}

	requisicao := protocol.Requisicao{
		Operacao: "criar_carona",
		Dados:    dadosJSON,
	}

	err = json.NewEncoder(conexao).Encode(requisicao)

	if err != nil {
		log.Fatalf("nao foi possivel enviar requisicao: %v", err)
	}

	var resposta protocol.Resposta
	err = json.NewDecoder(conexao).Decode(&resposta)
	if err != nil {
		log.Fatalf("nao foi possivel ler resposta: %v", err)
	}

	fmt.Println("Resposta do servidor:", resposta.Mensagem)
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

func lerRota(leitor *bufio.Reader) []string {
	origem := lerTexto(leitor, "Digite a cidade de origem da rota: ")
	rota := []string{origem}

	for {
		adiciona := lerTexto(leitor, "Deseja adicionar mais uma cidade à rota? (1) Sim (2) Não: ")
		if adiciona == "1" {
			novaCidade := lerTexto(leitor, "Digite a próxima cidade: ")
			rota = append(rota, novaCidade)
		} else if adiciona == "2" {
			if len(rota) < 2 {
				fmt.Println("A rota precisa ter ao menos duas cidades")
				continue
			}

			return rota
		} else {
			fmt.Println("Digite uma resposta válida")
		}
	}
}

func lerCapacidade(leitor *bufio.Reader) int {
	for {
		textoCapacidade := lerTexto(leitor, "Capacidade do veículo: ")
		capacidade, err := strconv.Atoi(textoCapacidade)
		if err != nil || capacidade <= 0 {
			fmt.Println("Digite uma capacidade inteira maior que zero")
			continue
		}

		return capacidade
	}
}

func lerHorarioSaida(leitor *bufio.Reader) string {
	for {
		data := lerTexto(leitor, "Data de saída (dd/mm/aaaa): ")
		hora := lerTexto(leitor, "Horário de saída (hh:mm): ")

		horarioSaida, err := converterDataHoraParaRFC3339(data, hora)
		if err != nil {
			fmt.Println("Data ou horário inválidos; use dd/mm/aaaa e hh:mm")
			continue
		}

		err = validarHorarioFuturo(horarioSaida, time.Now())
		if err != nil {
			fmt.Println(err)
			continue
		}

		return horarioSaida
	}
}

func converterDataHoraParaRFC3339(data string, hora string) (string, error) {
	fusoBrasil := time.FixedZone("BRT", -3*60*60)
	horarioSaida, err := time.ParseInLocation(
		"02/01/2006 15:04",
		data+" "+hora,
		fusoBrasil,
	)
	if err != nil {
		return "", err
	}

	return horarioSaida.Format(time.RFC3339), nil
}

func validarHorarioFuturo(horarioRFC3339 string, agora time.Time) error {
	horarioSaida, err := time.Parse(time.RFC3339, horarioRFC3339)
	if err != nil {
		return err
	}

	if !horarioSaida.After(agora) {
		return fmt.Errorf("o horário de saída precisa estar no futuro")
	}

	return nil
}

func lerPrecosCentavos(leitor *bufio.Reader, rota []string) []int64 {
	precosCentavos := make([]int64, 0, len(rota)-1)

	for i := 0; i < len(rota)-1; i++ {
		pergunta := fmt.Sprintf(
			"Preço de %s até %s em reais (ex.: 15,50): ",
			rota[i],
			rota[i+1],
		)

		for {
			textoPreco := lerTexto(leitor, pergunta)
			preco, err := converterReaisParaCentavos(textoPreco)
			if err != nil {
				fmt.Println("Preço inválido:", err)
				continue
			}

			precosCentavos = append(precosCentavos, preco)
			break
		}
	}

	return precosCentavos
}

func converterReaisParaCentavos(textoPreco string) (int64, error) {
	valor := strings.TrimSpace(textoPreco)
	partes := strings.Split(strings.ReplaceAll(valor, ",", "."), ".")

	if len(partes) > 2 || partes[0] == "" {
		return 0, fmt.Errorf("use um valor como 15,50")
	}

	reais, err := strconv.ParseInt(partes[0], 10, 64)
	if err != nil || reais < 0 {
		return 0, fmt.Errorf("use um valor não negativo como 15,50")
	}

	centavos := int64(0)
	if len(partes) == 2 {
		if len(partes[1]) == 0 || len(partes[1]) > 2 {
			return 0, fmt.Errorf("use no máximo duas casas decimais")
		}

		centavos, err = strconv.ParseInt(partes[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("use um valor como 15,50")
		}

		if len(partes[1]) == 1 {
			centavos *= 10
		}
	}

	return reais*100 + centavos, nil
}
