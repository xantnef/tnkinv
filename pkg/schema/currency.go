package schema

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"../aux"
)

const (
	FigiUSD = "BBG0013HGFT4"
)

/* const */
var Currencies = aux.NewList(
	"USD",
	"RUB",
	"EUR",
)

var CurrenciesOrdered = [...]string{"USD", "EUR", "RUB"}

type CValue struct {
	Currency string
	Value    float64
}

func NewCValue(val float64, currency string) CValue {
	if !Currencies.Has(currency) {
		log.Fatalf("unknown currency %s", currency)
	}

	return CValue{
		Currency: currency,
		Value:    val,
	}
}

func (cv CValue) Mult(m float64) CValue {
	return CValue{
		Currency: cv.Currency,
		Value:    cv.Value * m,
	}
}

func (cv CValue) Div(m float64) CValue {
	return cv.Mult(1 / m)
}

func (cv CValue) String() string {
	return fmt.Sprintf("{%s %.2f}", cv.Currency, cv.Value)
}

func (cv CValue) Copy() *CValue {
	copy := cv
	return &copy
}
