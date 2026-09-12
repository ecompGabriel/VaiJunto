# Protocolo VaiJunto — rascunho inicial

## Decisoes da primeira meta

- Transporte: TCP diretamente com o pacote `net` do Go.
- Codificacao: UTF-8.
- Representacao: JSON.
- Enquadramento: uma mensagem JSON por linha, terminada em `\n`.
- Cada requisicao tera uma operacao.
- Cada resposta tera uma mensagem.

## Operacao implementada nesta etapa

- `ping`: verificar a comunicacao cliente-servidor.

## Formato das mensagens

Uma requisicao valida possui uma operacao:

```json
{"operacao":"ping"}
```

O servidor responde com uma mensagem:

```json
{"mensagem":"pong"}
```

## Pendencias para decidir antes da busca entre caronas

- Regras de horario e tempo minimo de conexao entre caronas.
- Limite de conexoes e de resultados por busca.
- Ordenacao dos itinerarios.
- Autenticacao e duracao de sessao.
