# Protocolo inicial

Nesta etapa, clientes e servidor se comunicam por TCP usando mensagens JSON.
Cada mensagem é enviada em uma linha, permitindo que o servidor delimite as
requisições no fluxo TCP.

## Operação de teste

O cliente envia:

```json
{"operacao":"ping"}
```

O servidor responde:

```json
{"mensagem":"pong"}
```

Esse protocolo será evoluído para operações reais, como cadastro de carona,
busca de itinerários e confirmação de reservas.
