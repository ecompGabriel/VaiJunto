package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"vaijunto/internal/clienttcp"
	"vaijunto/internal/protocol"
)

func main() {
	leitor := bufio.NewReader(os.Stdin)

	cliente, err := clienttcp.Conectar()
	if err != nil {
		log.Fatalf("não foi possível conectar ao servidor: %v", err)
	}
	defer cliente.Fechar()

	if !autenticar(leitor, cliente, protocol.PerfilMotorista) {
		return
	}

	for {
		fmt.Println("\nVaiJunto - Motorista")
		fmt.Println("1 - Publicar carona")
		fmt.Println("2 - Consultar caronas e passageiros")
		fmt.Println("0 - Sair")

		opcao := lerTexto(leitor, "Escolha: ")
		switch opcao {
		case "1":
			publicarCarona(leitor, cliente)
		case "2":
			consultarCaronas(cliente)
		case "0":
			fmt.Println("Até logo!")
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}

func autenticar(leitor *bufio.Reader, cliente *clienttcp.Cliente, perfil string) bool {
	for {
		fmt.Println("\nAutenticação")
		fmt.Println("1 - Entrar")
		fmt.Println("2 - Cadastrar e entrar")
		fmt.Println("0 - Sair")

		opcao := lerTexto(leitor, "Escolha: ")
		if opcao == "0" {
			fmt.Println("Até logo!")
			return false
		}
		if opcao != "1" && opcao != "2" {
			fmt.Println("Opção inválida.")
			continue
		}

		usuarioID := lerTexto(leitor, "ID do usuário: ")
		senha := lerTexto(leitor, "Senha (mínimo de 4 caracteres): ")

		if opcao == "2" {
			resposta, err := cliente.RegistrarUsuario(usuarioID, senha, perfil)
			if err != nil {
				fmt.Println("Erro de comunicação ao cadastrar:", err)
				continue
			}
			if !resposta.Sucesso {
				fmt.Println("Cadastro recusado:", resposta.Mensagem)
				continue
			}
		}

		resposta, err := cliente.IniciarSessao(usuarioID, senha)
		if err != nil {
			fmt.Println("Erro de comunicação ao entrar:", err)
			continue
		}
		if !resposta.Sucesso {
			fmt.Println("Autenticação recusada:", resposta.Mensagem)
			continue
		}

		fmt.Println("Sessão iniciada com sucesso.")
		return true
	}
}

func publicarCarona(leitor *bufio.Reader, cliente *clienttcp.Cliente) {
	id := lerTexto(leitor, "ID da carona: ")
	horarioSaida := lerHorarioSaida(leitor)
	rota := lerRota(leitor)
	capacidade := lerCapacidade(leitor)
	precosCentavos := lerPrecosCentavos(leitor, rota)

	dados := protocol.CriarCarona{
		ID:             id,
		HorarioSaida:   horarioSaida,
		Rota:           rota,
		Capacidade:     capacidade,
		PrecosCentavos: precosCentavos,
	}

	resposta, err := cliente.Enviar("criar_carona", dados, nil)
	if err != nil {
		fmt.Println("Erro de comunicação:", err)
		return
	}

	fmt.Println(resposta.Mensagem)
}

func consultarCaronas(cliente *clienttcp.Cliente) {
	var caronas []protocol.CaronaDoMotorista
	resposta, err := cliente.Enviar("listar_caronas_motorista", struct{}{}, &caronas)
	if err != nil {
		fmt.Println("Erro de comunicação:", err)
		return
	}
	if !resposta.Sucesso {
		fmt.Println("Consulta recusada:", resposta.Mensagem)
		return
	}

	if len(caronas) == 0 {
		fmt.Println("Você ainda não publicou caronas.")
		return
	}

	for _, carona := range caronas {
		fmt.Printf("\nCarona %s\n", carona.ID)
		for _, trecho := range carona.Trechos {
			fmt.Printf(
				"- Trecho %d: %s → %s - %d/%d vaga(s) disponível(is)\n",
				trecho.Ordem,
				trecho.Origem,
				trecho.Destino,
				trecho.AssentosDisponiveis,
				trecho.Capacidade,
			)

			if len(trecho.Passageiros) == 0 {
				fmt.Println("  Nenhum passageiro confirmado.")
				continue
			}

			for _, passageiro := range trecho.Passageiros {
				fmt.Printf(
					"  Passageiro %s: %d assento(s)\n",
					passageiro.PassageiroID,
					passageiro.QuantidadeAssentos,
				)
			}
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
