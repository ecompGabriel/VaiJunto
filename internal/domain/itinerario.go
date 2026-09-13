package domain

type Itinerario struct {
	Origem             string
	Destino            string
	Trechos            []Trecho
	PrecoTotalCentavos int64
}

func (i Itinerario) calcularPrecoTotal() int64 {
	var precoTotal int64

	for _, trecho := range i.Trechos {
		precoTotal += trecho.PrecoCentavos
	}
	return precoTotal
}

func (i Itinerario) temVagas(quantidade int) bool {

	for _, trecho := range i.Trechos {
		if !trecho.TemVagas(quantidade) {
			return false
		}
	}
	return true
}
