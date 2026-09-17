package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"vaijunto/internal/auth"
	"vaijunto/internal/domain"
	"vaijunto/internal/protocol"
	"vaijunto/internal/search"
	"vaijunto/internal/store"
)

const (
	enderecoPadrao          = ":8080"
	tamanhoMaximoMensagem   = 64 * 1024
	tempoMaximoSemAtividade = 10 * time.Minute
	tempoMaximoEscrita      = 10 * time.Second
)

type sessaoConexao struct {
	// A sessão pertence a uma conexão TCP; ela desaparece quando a conexão fecha.
	UsuarioID string
	Perfil    string
}

func main() {
	endereco := os.Getenv("VAIJUNTO_LISTEN")
	if endereco == "" {
		endereco = enderecoPadrao
	}

	listener, err := net.Listen("tcp", endereco)
	if err != nil {
		log.Fatalf("não foi possível iniciar o servidor: %v", err)
	}
	defer listener.Close()

	catalogo := store.NovoCatalogoCaronas()
	usuarios := auth.NovoGerenciadorUsuarios()
	log.Printf("servidor ouvindo em %s", endereco)

	for {
		conexao, err := listener.Accept()
		if err != nil {
			log.Printf("erro ao aceitar conexão: %v", err)
			continue
		}

		// Cada cliente é atendido em sua própria goroutine. O catálogo sincroniza
		// o estado compartilhado entre todas elas.
		go tratarConexao(conexao, catalogo, usuarios)
	}
}

func tratarConexao(
	conexao net.Conn,
	catalogo *store.CatalogoCaronas,
	usuarios *auth.GerenciadorUsuarios,
) {
	defer conexao.Close()

	// O protocolo usa um JSON UTF-8 por linha. Scanner lida com leituras TCP
	// parciais e com várias mensagens recebidas no mesmo fluxo de bytes.
	leitor := bufio.NewScanner(conexao)
	leitor.Buffer(make([]byte, 4096), tamanhoMaximoMensagem)
	sessao := sessaoConexao{}

	for {
		// Um cliente ocioso não pode manter a goroutine presa indefinidamente.
		err := conexao.SetReadDeadline(time.Now().Add(tempoMaximoSemAtividade))
		if err != nil {
			return
		}

		if !leitor.Scan() {
			break
		}

		// A decodificação estrita rejeita JSON malformado e campos desconhecidos
		// sem derrubar o servidor nem afetar outras conexões.
		var requisicao protocol.Requisicao
		err = protocol.DecodificarEstrito(leitor.Bytes(), &requisicao)
		if err != nil {
			enviarResposta(conexao, respostaErro("", "mensagem_invalida", "requisição JSON inválida"))
			continue
		}

		resposta := tratarRequisicao(requisicao, &sessao, catalogo, usuarios)
		if !enviarResposta(conexao, resposta) {
			return
		}
	}

	erroLeitura := leitor.Err()
	if erroLeitura != nil {
		log.Printf("conexão encerrada com erro: %v", erroLeitura)
	}
}

func tratarRequisicao(
	requisicao protocol.Requisicao,
	sessao *sessaoConexao,
	catalogo *store.CatalogoCaronas,
	usuarios *auth.GerenciadorUsuarios,
) protocol.Resposta {
	// Versão e ID fazem parte de todas as operações para compatibilidade e para
	// correlacionar uma resposta com a requisição que a originou.
	if requisicao.Versao != protocol.VersaoAtual {
		return respostaErro(requisicao.ID, "versao_incompativel", "versão de protocolo incompatível")
	}

	if strings.TrimSpace(requisicao.ID) == "" {
		return respostaErro("", "id_obrigatorio", "o ID de correlação é obrigatório")
	}

	if requisicao.Operacao == "ping" {
		return respostaSucesso(requisicao.ID, "pong", nil)
	}

	if requisicao.Operacao == "registrar_usuario" {
		return tratarRegistrarUsuario(requisicao, usuarios)
	}

	if requisicao.Operacao == "iniciar_sessao" {
		return tratarIniciarSessao(requisicao, sessao, usuarios)
	}

	// Operações de negócio só são aceitas depois da autenticação nesta conexão.
	if sessao.UsuarioID == "" {
		return respostaErro(requisicao.ID, "sessao_obrigatoria", "inicie uma sessão antes desta operação")
	}

	switch requisicao.Operacao {
	case "criar_carona":
		if sessao.Perfil != protocol.PerfilMotorista {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do motorista")
		}
		return tratarCriarCarona(requisicao, catalogo, sessao.UsuarioID)

	case "listar_caronas_motorista":
		if sessao.Perfil != protocol.PerfilMotorista {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do motorista")
		}
		return tratarListarCaronasMotorista(requisicao, catalogo, sessao.UsuarioID)

	case "cancelar_carona":
		if sessao.Perfil != protocol.PerfilMotorista {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do motorista")
		}
		return tratarCancelarCarona(requisicao, catalogo, sessao.UsuarioID)

	case "buscar_itinerarios":
		if sessao.Perfil != protocol.PerfilPassageiro {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do passageiro")
		}
		return tratarBuscarItinerarios(requisicao, catalogo)

	case "confirmar_reserva":
		if sessao.Perfil != protocol.PerfilPassageiro {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do passageiro")
		}
		return tratarConfirmarReserva(requisicao, catalogo, sessao.UsuarioID)

	case "consultar_reservas":
		if sessao.Perfil != protocol.PerfilPassageiro {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do passageiro")
		}
		return tratarConsultarReservas(requisicao, catalogo, sessao.UsuarioID)

	case "cancelar_reserva":
		if sessao.Perfil != protocol.PerfilPassageiro {
			return respostaErro(requisicao.ID, "perfil_nao_autorizado", "operação exclusiva do passageiro")
		}
		return tratarCancelarReserva(requisicao, catalogo, sessao.UsuarioID)

	default:
		return respostaErro(requisicao.ID, "operacao_desconhecida", "operação desconhecida")
	}
}

func tratarRegistrarUsuario(
	requisicao protocol.Requisicao,
	usuarios *auth.GerenciadorUsuarios,
) protocol.Resposta {
	var dados protocol.RegistrarUsuario
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados de cadastro inválidos")
	}

	err = usuarios.Registrar(dados.UsuarioID, dados.Senha, dados.Perfil)
	if err != nil {
		return respostaErro(requisicao.ID, "usuario_nao_registrado", err.Error())
	}

	return respostaSucesso(requisicao.ID, "usuário cadastrado com sucesso", nil)
}

func tratarIniciarSessao(
	requisicao protocol.Requisicao,
	sessao *sessaoConexao,
	usuarios *auth.GerenciadorUsuarios,
) protocol.Resposta {
	if sessao.UsuarioID != "" {
		return respostaErro(requisicao.ID, "sessao_ja_iniciada", "esta conexão já possui uma sessão")
	}

	var dados protocol.IniciarSessao
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da sessão inválidos")
	}

	perfil, err := usuarios.Autenticar(dados.UsuarioID, dados.Senha)
	if err != nil {
		return respostaErro(requisicao.ID, "autenticacao_falhou", err.Error())
	}

	sessao.UsuarioID = strings.TrimSpace(dados.UsuarioID)
	sessao.Perfil = perfil

	return respostaSucesso(requisicao.ID, "sessão iniciada", protocol.SessaoIniciada{
		UsuarioID: sessao.UsuarioID,
		Perfil:    sessao.Perfil,
	})
}

func tratarCriarCarona(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	motoristaID string,
) protocol.Resposta {
	var dados protocol.CriarCarona
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da carona inválidos")
	}

	horarioSaida, err := time.Parse(time.RFC3339, dados.HorarioSaida)
	if err != nil {
		return respostaErro(requisicao.ID, "horario_invalido", "horário de saída inválido")
	}

	carona, err := domain.NovaCarona(
		dados.ID,
		motoristaID,
		horarioSaida,
		dados.Rota,
		dados.Capacidade,
		dados.PrecosCentavos,
		dados.DuracoesMinutos,
	)
	if err != nil {
		return respostaErro(requisicao.ID, "carona_invalida", err.Error())
	}

	err = catalogo.Adicionar(carona)
	if err != nil {
		return respostaErro(requisicao.ID, "carona_nao_criada", err.Error())
	}

	return respostaSucesso(requisicao.ID, "carona criada com sucesso", nil)
}

func tratarBuscarItinerarios(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
) protocol.Resposta {
	var dados protocol.BuscarItinerarios
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da busca inválidos")
	}
	dataDesejada, err := time.Parse(time.RFC3339, dados.DataDesejada)
	if err != nil {
		return respostaErro(requisicao.ID, "data_invalida", "data desejada inválida")
	}

	itinerarios, err := search.BuscarItinerarios(
		catalogo.ListarTrechos(),
		dados.Origem,
		dados.Destino,
		dados.QuantidadeAssentos,
		dataDesejada,
	)
	if err != nil {
		return respostaErro(requisicao.ID, "busca_invalida", err.Error())
	}

	resultados := make([]protocol.ItinerarioEncontrado, 0, len(itinerarios))
	for _, itinerario := range itinerarios {
		trechos := make([]protocol.TrechoEncontrado, 0, len(itinerario.Trechos))
		for _, trecho := range itinerario.Trechos {
			trechos = append(trechos, protocol.TrechoEncontrado{
				CaronaID:            trecho.CaronaID,
				Ordem:               trecho.Ordem,
				Origem:              trecho.Origem,
				Destino:             trecho.Destino,
				PrecoCentavos:       trecho.PrecoCentavos,
				AssentosDisponiveis: trecho.AssentosDisponiveis,
				HorarioSaida:        trecho.HorarioSaida.Format(time.RFC3339),
				HorarioChegada:      trecho.HorarioChegada.Format(time.RFC3339),
			})
		}

		resultados = append(resultados, protocol.ItinerarioEncontrado{
			Origem:             itinerario.Origem,
			Destino:            itinerario.Destino,
			Trechos:            trechos,
			PrecoTotalCentavos: itinerario.PrecoTotalCentavos,
		})
	}

	mensagem := "itinerários encontrados"
	if len(resultados) == 0 {
		mensagem = "nenhum itinerário encontrado"
	}

	return respostaSucesso(requisicao.ID, mensagem, resultados)
}

func tratarConfirmarReserva(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	passageiroID string,
) protocol.Resposta {
	var dados protocol.ConfirmarReserva
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da reserva inválidos")
	}

	referencias := make([]domain.ReferenciaTrecho, 0, len(dados.Trechos))
	for _, trecho := range dados.Trechos {
		referencias = append(referencias, domain.ReferenciaTrecho{
			CaronaID: trecho.CaronaID,
			Ordem:    trecho.Ordem,
		})
	}

	reserva, err := catalogo.ConfirmarReserva(
		dados.IDReserva,
		passageiroID,
		dados.QuantidadeAssentos,
		referencias,
	)
	if err != nil {
		return respostaErro(requisicao.ID, "reserva_nao_confirmada", err.Error())
	}

	return respostaSucesso(requisicao.ID, "reserva confirmada com sucesso", reservaParaProtocolo(*reserva))
}

func tratarConsultarReservas(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	passageiroID string,
) protocol.Resposta {
	var dados struct{}
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da consulta inválidos")
	}

	reservasDominio := catalogo.ListarReservasDoPassageiro(passageiroID)
	reservas := make([]protocol.Reserva, 0, len(reservasDominio))
	for _, reserva := range reservasDominio {
		reservas = append(reservas, reservaParaProtocolo(reserva))
	}

	return respostaSucesso(requisicao.ID, "reservas consultadas", reservas)
}

func tratarCancelarReserva(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	passageiroID string,
) protocol.Resposta {
	var dados protocol.CancelarReserva
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados do cancelamento inválidos")
	}

	err = catalogo.CancelarReserva(dados.IDReserva, passageiroID)
	if err != nil {
		return respostaErro(requisicao.ID, "reserva_nao_cancelada", err.Error())
	}

	return respostaSucesso(requisicao.ID, "reserva cancelada com sucesso", nil)
}

func tratarCancelarCarona(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	motoristaID string,
) protocol.Resposta {
	var dados protocol.CancelarCarona
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados do cancelamento inválidos")
	}

	err = catalogo.CancelarCarona(dados.IDCarona, motoristaID)
	if err != nil {
		return respostaErro(requisicao.ID, "carona_nao_cancelada", err.Error())
	}

	return respostaSucesso(requisicao.ID, "carona cancelada com sucesso", nil)
}

func tratarListarCaronasMotorista(
	requisicao protocol.Requisicao,
	catalogo *store.CatalogoCaronas,
	motoristaID string,
) protocol.Resposta {
	var dados struct{}
	err := protocol.DecodificarEstrito(requisicao.Dados, &dados)
	if err != nil {
		return respostaErro(requisicao.ID, "dados_invalidos", "dados da consulta inválidos")
	}

	situacoes := catalogo.ListarCaronasDoMotorista(motoristaID)
	caronas := make([]protocol.CaronaDoMotorista, 0, len(situacoes))

	for _, situacao := range situacoes {
		trechos := make([]protocol.TrechoDoMotorista, 0, len(situacao.Trechos))
		for _, situacaoTrecho := range situacao.Trechos {
			passageiros := make([]protocol.PassageiroConfirmado, 0, len(situacaoTrecho.Passageiros))
			for _, passageiro := range situacaoTrecho.Passageiros {
				passageiros = append(passageiros, protocol.PassageiroConfirmado{
					PassageiroID:       passageiro.PassageiroID,
					QuantidadeAssentos: passageiro.QuantidadeAssentos,
				})
			}

			trecho := situacaoTrecho.Trecho
			trechos = append(trechos, protocol.TrechoDoMotorista{
				Ordem:               trecho.Ordem,
				Origem:              trecho.Origem,
				Destino:             trecho.Destino,
				Capacidade:          trecho.Capacidade,
				AssentosDisponiveis: trecho.AssentosDisponiveis,
				HorarioSaida:        trecho.HorarioSaida.Format(time.RFC3339),
				HorarioChegada:      trecho.HorarioChegada.Format(time.RFC3339),
				Passageiros:         passageiros,
			})
		}

		caronas = append(caronas, protocol.CaronaDoMotorista{
			ID:        situacao.ID,
			Cancelada: situacao.Cancelada,
			Trechos:   trechos,
		})
	}

	return respostaSucesso(requisicao.ID, "caronas consultadas", caronas)
}

func reservaParaProtocolo(reserva domain.Reserva) protocol.Reserva {
	referenciasDominio := reserva.Trechos()
	referencias := make([]protocol.ReferenciaTrecho, 0, len(referenciasDominio))

	for _, referencia := range referenciasDominio {
		referencias = append(referencias, protocol.ReferenciaTrecho{
			CaronaID: referencia.CaronaID,
			Ordem:    referencia.Ordem,
		})
	}

	return protocol.Reserva{
		ID:                 reserva.ID,
		PassageiroID:       reserva.PassageiroID,
		QuantidadeAssentos: reserva.QuantidadeAssentos,
		Status:             string(reserva.Status),
		Trechos:            referencias,
	}
}

func respostaSucesso(id string, mensagem string, dados any) protocol.Resposta {
	resposta := protocol.Resposta{
		Versao:   protocol.VersaoAtual,
		ID:       id,
		Sucesso:  true,
		Codigo:   "ok",
		Mensagem: mensagem,
	}

	if dados == nil {
		return resposta
	}

	dadosJSON, err := json.Marshal(dados)
	if err != nil {
		return respostaErro(id, "erro_interno", "não foi possível preparar a resposta")
	}

	resposta.Dados = dadosJSON
	return resposta
}

func respostaErro(id string, codigo string, mensagem string) protocol.Resposta {
	return protocol.Resposta{
		Versao:   protocol.VersaoAtual,
		ID:       id,
		Sucesso:  false,
		Codigo:   codigo,
		Mensagem: mensagem,
	}
}

func enviarResposta(conexao net.Conn, resposta protocol.Resposta) bool {
	// O limite de escrita impede que um cliente que parou de ler bloqueie o
	// atendimento da sua própria conexão para sempre.
	err := conexao.SetWriteDeadline(time.Now().Add(tempoMaximoEscrita))
	if err != nil {
		return false
	}

	err = json.NewEncoder(conexao).Encode(resposta)
	return err == nil
}
