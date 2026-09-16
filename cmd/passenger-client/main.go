package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
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

	if !autenticar(leitor, cliente, protocol.PerfilPassageiro) {
		return
	}

	for {
		fmt.Println("\nVaiJunto - Passageiro")
		fmt.Println("1 - Buscar itinerários")
		fmt.Println("2 - Consultar minhas reservas")
		fmt.Println("3 - Cancelar reserva")
		fmt.Println("0 - Sair")

		opcao := lerTexto(leitor, "Escolha: ")
		switch opcao {
		case "1":
			buscarEReservar(leitor, cliente)
		case "2":
			consultarReservas(cliente)
		case "3":
			cancelarReserva(leitor, cliente)
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

func buscarEReservar(leitor *bufio.Reader, cliente *clienttcp.Cliente) {
	origem := lerTexto(leitor, "Cidade de origem: ")
	destino := lerTexto(leitor, "Cidade de destino: ")
	quantidadeAssentos := lerInteiroPositivo(leitor, "Quantidade de assentos: ")

	dados := protocol.BuscarItinerarios{
		Origem:             origem,
		Destino:            destino,
		QuantidadeAssentos: quantidadeAssentos,
	}

	var itinerarios []protocol.ItinerarioEncontrado
	resposta, err := cliente.Enviar("buscar_itinerarios", dados, &itinerarios)
	if err != nil {
		fmt.Println("Erro de comunicação:", err)
		return
	}
	if !resposta.Sucesso {
		fmt.Println("Busca recusada:", resposta.Mensagem)
		return
	}

	if len(itinerarios) == 0 {
		fmt.Println("Nenhum itinerário encontrado.")
		return
	}

	mostrarItinerarios(itinerarios)

	for {
		escolha := lerInteiroNaoNegativo(
			leitor,
			"Digite o número da opção para reservar ou 0 para voltar: ",
		)
		if escolha == 0 {
			return
		}
		if escolha > len(itinerarios) {
			fmt.Println("Opção inexistente.")
			continue
		}

		itinerario := itinerarios[escolha-1]
		referencias := make([]protocol.ReferenciaTrecho, 0, len(itinerario.Trechos))
		for _, trecho := range itinerario.Trechos {
			referencias = append(referencias, protocol.ReferenciaTrecho{
				CaronaID: trecho.CaronaID,
				Ordem:    trecho.Ordem,
			})
		}

		confirmar := protocol.ConfirmarReserva{
			IDReserva:          novoIDReserva(),
			QuantidadeAssentos: quantidadeAssentos,
			Trechos:            referencias,
		}

		var reserva protocol.Reserva
		resposta, err = cliente.Enviar("confirmar_reserva", confirmar, &reserva)
		if err != nil {
			fmt.Println("Erro de comunicação:", err)
			return
		}
		if !resposta.Sucesso {
			fmt.Println("Reserva recusada:", resposta.Mensagem)
			return
		}

		fmt.Printf("Reserva %s confirmada para %d assento(s).\n", reserva.ID, reserva.QuantidadeAssentos)
		return
	}
}

func consultarReservas(cliente *clienttcp.Cliente) {
	var reservas []protocol.Reserva
	resposta, err := cliente.Enviar("consultar_reservas", struct{}{}, &reservas)
	if err != nil {
		fmt.Println("Erro de comunicação:", err)
		return
	}
	if !resposta.Sucesso {
		fmt.Println("Consulta recusada:", resposta.Mensagem)
		return
	}

	if len(reservas) == 0 {
		fmt.Println("Você ainda não possui reservas.")
		return
	}

	for _, reserva := range reservas {
		fmt.Printf(
			"\nReserva %s - %s - %d assento(s)\n",
			reserva.ID,
			reserva.Status,
			reserva.QuantidadeAssentos,
		)
		for _, trecho := range reserva.Trechos {
			fmt.Printf("- Carona %s, trecho %d\n", trecho.CaronaID, trecho.Ordem)
		}
	}
}

func cancelarReserva(leitor *bufio.Reader, cliente *clienttcp.Cliente) {
	idReserva := lerTexto(leitor, "ID da reserva a cancelar: ")
	dados := protocol.CancelarReserva{IDReserva: idReserva}

	resposta, err := cliente.Enviar("cancelar_reserva", dados, nil)
	if err != nil {
		fmt.Println("Erro de comunicação:", err)
		return
	}

	fmt.Println(resposta.Mensagem)
}

func mostrarItinerarios(itinerarios []protocol.ItinerarioEncontrado) {
	for indice, itinerario := range itinerarios {
		fmt.Printf("\nOpção %d - %s até %s\n", indice+1, itinerario.Origem, itinerario.Destino)
		fmt.Printf(
			"Preço total: R$ %d,%02d\n",
			itinerario.PrecoTotalCentavos/100,
			itinerario.PrecoTotalCentavos%100,
		)

		for _, trecho := range itinerario.Trechos {
			fmt.Printf(
				"- %s → %s - R$ %d,%02d - %d vaga(s)\n",
				trecho.Origem,
				trecho.Destino,
				trecho.PrecoCentavos/100,
				trecho.PrecoCentavos%100,
				trecho.AssentosDisponiveis,
			)
		}
	}
}

func novoIDReserva() string {
	bytesAleatorios := make([]byte, 8)
	_, err := rand.Read(bytesAleatorios)
	if err == nil {
		return "reserva-" + hex.EncodeToString(bytesAleatorios)
	}

	return fmt.Sprintf("reserva-%d", time.Now().UnixNano())
}

func lerInteiroPositivo(leitor *bufio.Reader, pergunta string) int {
	for {
		valor := lerInteiroNaoNegativo(leitor, pergunta)
		if valor > 0 {
			return valor
		}

		fmt.Println("Digite um número maior que zero.")
	}
}

func lerInteiroNaoNegativo(leitor *bufio.Reader, pergunta string) int {
	for {
		texto := lerTexto(leitor, pergunta)
		valor, err := strconv.Atoi(texto)
		if err != nil || valor < 0 {
			fmt.Println("Digite um número inteiro não negativo.")
			continue
		}

		return valor
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
			fmt.Println("Este campo não pode ficar vazio.")
			continue
		}

		return texto
	}
}
