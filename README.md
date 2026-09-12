# VaiJunto

Projeto da disciplina TEC502 sobre um sistema de caronas compartilhadas.

## Meta atual

Definir um protocolo inicial de requisicoes e respostas em JSON sobre TCP e
iniciar a modelagem de caronas, trechos e itinerarios.

Nesta etapa, o servidor e os dois clientes testam o protocolo por meio da
operacao `ping`.

## Estrutura inicial

- `cmd/server`: ponto de entrada do servidor TCP.
- `cmd/driver-client`: cliente inicial do motorista.
- `cmd/passenger-client`: cliente inicial do passageiro.
- `internal/protocol`: representacao de requisicoes e respostas.
- `internal/domain`: modelos de carona, trecho e itinerario.
- `docs`: especificacao do protocolo.

## Executar a meta atual

Em um terminal, inicie o servidor:

```bash
go run ./cmd/server
```

Em outros terminais, execute os clientes:

```bash
go run ./cmd/driver-client
go run ./cmd/passenger-client
```

Nesta primeira versao, os clientes conectam em `127.0.0.1:8080`, isto e, no
mesmo computador. O IP de outro computador sera configurado depois.
