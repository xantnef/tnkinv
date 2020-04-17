package schema

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Pricef0 func() float64
type Pricef1 func(figi string) float64
type Pricef2 func(figi string, t time.Time) float64

func PriceCurry0(f1 Pricef1, figi string) Pricef0 {
	return func() float64 {
		return f1(figi)
	}
}

// =============================================================================

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
		cv := NewCValue(0, "RUB")
		m["all"] = &cv
	})

	return b
}

func (b *Balance) Get(currency string) CValue {
	return NewCValue(
		b.Assets[currency].Value-b.Payins[currency].Value,
		b.Assets[currency].Currency)
}

func (b *Balance) Foreach(f func(string, CurMap)) {
	names := []string{"Payins", "Assets", "Commissions"}
	for i, m := range []CurMap{b.Payins, b.Assets, b.Commissions} {
		f(names[i], m)
	}
}

func (b *Balance) Copy() *Balance {
	copy := NewBalance()

	for cur := range b.Payins {
		copy.Payins[cur] = b.Payins[cur].Copy()
		copy.Assets[cur] = b.Assets[cur].Copy()
		copy.Commissions[cur] = b.Commissions[cur].Copy()
	}

	return copy
}

func (b *Balance) Add(b2 Balance) {
	for cur := range b.Payins {
		b.Payins[cur].Value += b2.Payins[cur].Value
		b.Assets[cur].Value += b2.Assets[cur].Value
		b.Commissions[cur].Value += b2.Commissions[cur].Value
	}
}

func (b *Balance) CalcAllAssets(usd, eur float64) float64 {
	xchgrate := map[string]float64{
		"RUB": 1,
		"USD": usd,
		"EUR": eur,
	}

	b.Assets["all"].Value = 0
	for cur := range Currencies {
		b.Assets["all"].Value += b.Assets[cur].Value * xchgrate[cur]
	}
	return b.Assets["all"].Value
}

const (
	TableStyle = "table"
)

func (b Balance) ToString(prefix string, style string) (s string) {
	actualCurrencies := []string{}

	for _, cur := range CurrenciesOrdered {
		if b.Payins[cur].Value != 0 {
			actualCurrencies = append(actualCurrencies, cur)
		}
	}

	actualCurrencies = append(actualCurrencies, "all")

	if style == TableStyle {
		s += fmt.Sprintf("%s: ", prefix)
		for _, cur := range actualCurrencies {
			s += fmt.Sprintf("%f, %f, %f, ",
				b.Payins[cur].Value, b.Assets[cur].Value, b.Get(cur).Value)
		}
		s += fmt.Sprintln()
	} else {
		for _, cur := range actualCurrencies {
			s += fmt.Sprintf("%s: %s: %7.0f %7.0f %7.0f\n",
				prefix, cur,
				b.Payins[cur].Value, b.Assets[cur].Value, b.Get(cur).Value)
		}
		s += fmt.Sprintln()
	}

	return
}

// =============================================================================

type Deal struct {
	Date       time.Time
	Price      CValue
	Quantity   int
	Accrued    float64 // aka NKD
	Commission float64
}

func (deal Deal) String() string {
	return fmt.Sprintf(
		"%s: (%.2f x %d) = %s",
		deal.Date.Format("2006/01/02"),
		deal.Price.Value,
		deal.Quantity,
		deal.CValue())
}

func (deal Deal) Value() float64 {
	return deal.Price.Value*float64(deal.Quantity) + deal.Accrued
}

func (deal Deal) CValue() CValue {
	cv := deal.Price.Mult(float64(deal.Quantity))
	cv.Value += deal.Accrued
	return cv
}

// =============================================================================

func (op Operation) StringPretty() string {
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

func (op Operation) IsPayment() bool {
	// TODO PartRepayment doesnt really belong here.
	// make it a pseudo-deal?
	return op.OperationType == "Dividend" ||
		op.OperationType == "TaxDividend" ||
		op.OperationType == "Coupon" ||
		op.OperationType == "TaxCoupon" ||
		op.OperationType == "PartRepayment"
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
	AvgPrice CValue // TODO unused

	Balance     CValue
	Yield       float64
	YieldAnnual float64
	YieldMarket float64

	// aux and temporary
	SplitSells []*Deal
}

func (po *Portion) CheckNoSplitSells(ticker string) {
	if len(po.SplitSells) > 0 {
		// those split sells were indeed partial
		log.Errorf("%s: partial sells are not handled nicely yet %s",
			ticker, po.SplitSells)
		po.SplitSells = []*Deal{}
	}
}

func (po Portion) String() string {
	date := "----/--/--"
	if po.IsClosed {
		date = po.Close.Date.Format("2006/01/02")
	}

	return fmt.Sprintf(
		"%s: %s (%.1f%%, annual %.1f%%, market %.1f%%)",
		date, po.Balance, po.Yield, po.YieldAnnual, po.YieldMarket)
}

// =============================================================================
type RepaymentPoint struct {
	Time time.Time
	Mult float64
}

type PositionInfo struct {
	Ins Instrument

	Deals      []*Deal
	Dividends  []*Dividend
	Portions   []*Portion
	Repayments []*RepaymentPoint

	OpenQuantity int // TODO
	OpenSpent    float64
	OpenDeal     *Deal

	AccumulatedIncome CValue // TODO
}

func (pinfo *PositionInfo) IsClosed() bool {
	return pinfo.OpenDeal == nil
}

func (pinfo *PositionInfo) StringPretty() string {
	s := fmt.Sprintf("%s:", pinfo.Ins.Name)

	od := pinfo.OpenDeal
	if od != nil {
		s += fmt.Sprintf(" %s (%.2f x %d) +acc %v",
			od.CValue().Mult(-1.0),
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

// =============================================================================
