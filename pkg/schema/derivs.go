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

func (cv CValue) Copy() *CValue {
	copy := cv
	return &copy
}

// =============================================================================

type CurMap map[string]*CValue

type Balance struct {
	Commissions, Payins, Assets CurMap
}

func NewBalance() *Balance {
	b := &Balance{
		Commissions: make(CurMap),
		Payins:      make(CurMap),
		Assets:      make(CurMap),
	}

	b.Foreach(func(nm string, m CurMap) {
		for cur := range Currencies {
			cv := NewCValue(0, cur)
			m[cur] = &cv
		}
	})

	return b
}

func (b *Balance) Get(currency string) CValue {
	return NewCValue(
		b.Assets[currency].Value-b.Payins[currency].Value,
		currency)
}

func (b *Balance) Foreach(f func(string, CurMap)) {
	names := []string{"Payins", "Assets", "Commissions"}
	for i, m := range []CurMap{b.Payins, b.Assets, b.Commissions} {
		f(names[i], m)
	}
}

func (b *Balance) Copy() *Balance {
	copy := NewBalance()

	for cur := range Currencies {
		copy.Commissions[cur] = b.Commissions[cur].Copy()
		copy.Payins[cur] = b.Payins[cur].Copy()
		copy.Assets[cur] = b.Assets[cur].Copy()
	}

	return copy
}

func (b *Balance) GetTotal(usd, eur float64) (p, a, d float64) {
	getprice := func(cur string) float64 {
		if cur == "RUB" {
			return 1
		} else if cur == "EUR" {
			return eur
		} else if cur == "USD" {
			return usd
		}
		return 0
	}

	for cur := range Currencies {
		p += b.Payins[cur].Value * getprice(cur)
		a += b.Assets[cur].Value * getprice(cur)
		d += b.Get(cur).Value * getprice(cur)
	}
	return
}

// =============================================================================

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

// =============================================================================

type Dividend struct {
	Date  time.Time
	Value float64
}

// =============================================================================

type Portion struct {
	Buys  []*Deal
	Close *Deal

	IsClosed bool

	AvgDate  time.Time
	AvgPrice CValue // TODO

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

// =============================================================================

type PositionInfo struct {
	Figi   string
	Ticker string

	Deals     []*Deal
	Dividends []*Dividend
	Portions  []*Portion

	OpenQuantity int // TODO
	OpenSpent    float64
	OpenDeal     *Deal

	AccumulatedIncome CValue // TODO
}

func (pinfo *PositionInfo) IsClosed() bool {
	return pinfo.OpenDeal == nil
}

func (pinfo *PositionInfo) String() string {
	s := fmt.Sprintf("%s:", pinfo.Ticker)

	od := pinfo.OpenDeal
	if od != nil {
		s += fmt.Sprintf(" %s (%.2f x %d) +acc %v",
			od.Price.Mult(float64(-od.Quantity)),
			od.Price.Value, -od.Quantity, pinfo.AccumulatedIncome)
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
