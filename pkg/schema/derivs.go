package schema

import (
	"fmt"
	"log"
	"time"
)

const (
	FigiUSD = "BBG0013HGFT4" // ticker for USD buys
)

/* const */
var Currencies map[string]bool = map[string]bool{
	"USD": true,
	"RUB": true,
	"EUR": true,
}

type CValue struct {
	Currency string
	Value    float64
}

func NewCValue(val float64, currency string) CValue {
	if !Currencies[currency] {
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

type Deal struct {
	Date       time.Time
	Price      CValue
	Quantity   int
	Commission float64
}

func (deal Deal) String() string {
	return fmt.Sprintf(
		"%s: (%.2f x %d) = %s",
		deal.Date.Format("2006/01/02"),
		deal.Price.Value,
		deal.Quantity,
		deal.Price.Mult(float64(deal.Quantity)))
}

func (deal Deal) Value() float64 {
	return deal.Price.Value * float64(deal.Quantity)
}

type Dividend struct {
	Date  time.Time
	Value float64
}

type Portion struct {
	Buys  []*Deal
	Close *Deal

	IsClosed bool

	AvgDate  time.Time
	AvgPrice CValue

	Balance     CValue
	Yield       CValue
	YieldAnnual float64
}

func (po Portion) String() string {
	date := "----/--/--"
	if po.IsClosed {
		date = po.Close.Date.Format("2006/01/02")
	}

	return fmt.Sprintf(
		"%s: %s (%.1f%%, annual %.1f%%)",
		date, po.Balance, po.Yield.Value, po.YieldAnnual)
}

type PositionInfo struct {
	Figi     string
	Ticker   string
	IsClosed bool

	Deals     []*Deal
	Dividends []*Dividend
	Portions  []*Portion

	CurrentPrice      CValue
	Quantity          float64 // TODO remove
	AccumulatedIncome CValue
}

func (pinfo *PositionInfo) String() string {
	s := fmt.Sprintf("%s:", pinfo.Ticker)

	if !pinfo.IsClosed {
		s += fmt.Sprintf(" %s (%.2f x %.0f) +acc %v",
			pinfo.CurrentPrice.Mult(pinfo.Quantity),
			pinfo.CurrentPrice.Value, pinfo.Quantity, pinfo.AccumulatedIncome)
	}

	s += "\n" +
		"    deals:\n"

	for _, deal := range pinfo.Deals {
		s += "      " + deal.String() + "\n"
	}

	s += "    position stats:\n"

	for _, po := range pinfo.Portions {
		s += "      " + po.String() + "\n"
	}

	return s
}
