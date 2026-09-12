package protocol

type Requisicao struct {
	Operacao string `json:"operacao"`
}

type Resposta struct {
	Mensagem string `json:"mensagem"`
}
