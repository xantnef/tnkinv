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

var CurrenciesOrdered []string = []string{"USD", "EUR", "RUB"}

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

const (
	TableStyle = "table"
)

func (b Balance) ToString(t time.Time, usd, eur float64, style string) (s string) {
	actualCurrencies := []string{}

	for _, cur := range CurrenciesOrdered {
		if b.Payins[cur].Value != 0 {
			actualCurrencies = append(actualCurrencies, cur)
		}
	}

	p, a, d := b.GetTotal(usd, eur)

	if style == TableStyle {
		s += fmt.Sprintf("%s, ", t.Format("2006/01/02"))
		for _, cur := range actualCurrencies {
			s += fmt.Sprintf("%f, %f, %f, ",
				b.Payins[cur].Value, b.Assets[cur].Value, b.Get(cur).Value)
		}

		s += fmt.Sprintf("%f, %f, %f\n", p, a, d)

	} else {
		for _, cur := range actualCurrencies {
			s += fmt.Sprintf("%s, %s, %f, %f, %f\n",
				t.Format("2006/01/02"), cur,
				b.Payins[cur].Value, b.Assets[cur].Value, b.Get(cur).Value)
		}

		s += fmt.Sprintf("%s, %s, %f, %f, %f\n",
			t.Format("2006/01/02"), "tot", p, a, d)
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

func (op Operation) String() string {
	shortTick := op.Ticker
	if op.Figi == FigiUSD {
		shortTick = "USD"
	}

	return fmt.Sprintf(
		"%s: %-17s %-4s (%-7.2f x %-3d) = %s %-9.2f",
		op.Date, op.OperationType, shortTick, op.Price, op.Quantity,
		op.Currency, op.Payment)
}

func (op Operation) IsTrading() bool {
	return op.OperationType == "Buy" || op.OperationType == "BuyCard" ||
		op.OperationType == "Sell"
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
