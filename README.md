# VaiJunto

Projeto individual da disciplina TEC502 da UEFS. O VaiJunto é um sistema de
caronas compartilhadas de média e longa distância, desenvolvido em Go.

## Estado atual

O projeto possui o modelo inicial de carona, trecho e itinerário, com testes
automatizados de domínio. O servidor TCP e os clientes de motorista e
passageiro ainda usam uma comunicação inicial em JSON para validar a conexão.

## Estrutura

- `cmd/server`: servidor TCP.
- `cmd/driver-client`: cliente do motorista.
- `cmd/passenger-client`: cliente do passageiro.
- `internal/domain`: regras de carona, trechos e itinerários.
- `internal/protocol`: mensagens trocadas entre clientes e servidor.
- `docs`: documentação do projeto.

## Testes

Na raiz do projeto, execute:

```bash
go test ./...
```
