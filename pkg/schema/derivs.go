package schema

import (
	"fmt"
	"log"
	"time"
)

type CValue struct {
	Currency string
	Value    float64
}

func NewCValue(val float64, currency string) CValue {
	m := map[string]bool{
		"USD": true,
		"RUB": true,
	}

	if !m[currency] {
		log.Fatalf("unknown currency %s", currency)
	}

	return CValue{
		Currency: currency,
		Value:    val,
	}
}

func (cv *CValue) Mult(m float64) CValue {
	return CValue{
		Currency: cv.Currency,
		Value:    cv.Value * m,
	}
}

func (cv *CValue) Div(m float64) CValue {
	return cv.Mult(1 / m)
}

type Deal struct {
	Opened time.Time
	Closed time.Time

	IsSumdeal bool

	Price       CValue
	ClosedPrice CValue
	Quantity    int

	Yield       CValue
	YieldAnnual float64
}

func (deal *Deal) String() string {
	if deal.Quantity < 0 {
		return ""
	}

	fmap := map[bool]string{
		true:  "[SUM]",
		false: "     ",
	}

	return fmt.Sprintf(
		"%s: %s spent=%v yield=%.1f%% annual=%.1f%%",
		deal.Opened.Format("2006/01/02"),
		fmap[deal.IsSumdeal],
		deal.Price.Mult(float64(deal.Quantity)),
		deal.Yield.Value,
		deal.YieldAnnual)
}

type PositionInfo struct {
	Figi     string
	Ticker   string
	IsClosed bool

	Deals   []*Deal

	CurrentPrice      CValue
	Quantity          float64 // TODO filter currencies with float quantities out?
	AccumulatedIncome CValue
	YieldAnnual       float64
}

func (pinfo *PositionInfo) String() string {
	s := fmt.Sprintf(
		"%s: %v (%.0f x %.1f) +acc %v\n",
		pinfo.Ticker, pinfo.CurrentPrice.Mult(pinfo.Quantity),
		pinfo.Quantity, pinfo.CurrentPrice.Value, pinfo.AccumulatedIncome)

	for _, deal := range pinfo.Deals {
		s += "    " + deal.String() + "\n"
	}

	return s
}
