package domain

type Itinerario struct {
	Origem             string
	Destino            string
	Trechos            []Trecho
	PrecoTotalCentavos int64
}
