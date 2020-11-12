package schema

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	"../aux"
)

func balanceMaps() []string {
	return append(CurrenciesOrdered[:], "all")
}

type CurMap map[string]*CValue

func NewCurMap() CurMap {
	m := make(CurMap)
	for cur := range Currencies {
		cv := NewCValue(0, cur)
		m[cur] = &cv
	}
	cv := NewCValue(0, "RUB")
	m["all"] = &cv
	return m
}

func (m CurMap) CalcAll(usd, eur float64) float64 {
	xchgrate := map[string]float64{
		"RUB": 1,
		"USD": usd,
		"EUR": eur,
	}

	m["all"].Value = 0
	for cur := range Currencies {
		m["all"].Value += m[cur].Value * xchgrate[cur]
	}
	return m["all"].Value
}

func (m CurMap) Add(cv CValue) {
	m[cv.Currency].Value += cv.Value
}

func (m CurMap) String() string {
	s := ""
	for _, cur := range balanceMaps() {
		if m[cur].Value != 0 {
			if s != "" {
				s += ", "
			}
			s += fmt.Sprintf("{%s: %.2f}", cur, m[cur].Value)
		}
	}
	return s
}

// =============================================================================

type Balance struct {
	Commissions, Payins, Assets CurMap
	AvgDate                     time.Time
}

func NewBalance() *Balance {
	return &Balance{
		Commissions: NewCurMap(),
		Payins:      NewCurMap(),
		Assets:      NewCurMap(),
	}
}

func (b Balance) Get_(currency string) CValue { // TODO unused
	return NewCValue(
		b.Assets[currency].Value-b.Payins[currency].Value,
		b.Assets[currency].Currency)
}

func (b Balance) Foreach(f func(string, CurMap)) {
	names := []string{"Payins", "Assets", "Commissions"}
	for i, m := range []CurMap{b.Payins, b.Assets, b.Commissions} {
		f(names[i], m)
	}
}

func (b Balance) Copy() *Balance {
	copy := NewBalance()

	for _, cur := range balanceMaps() {
		copy.Payins[cur] = b.Payins[cur].Copy()
		copy.Assets[cur] = b.Assets[cur].Copy()
		copy.Commissions[cur] = b.Commissions[cur].Copy()
	}

	return copy
}

func (b Balance) hasPayins() bool {
	for _, cv := range b.Payins {
		if cv.Value > 0 {
			return true
		}
	}
	return false
}

func (b *Balance) Add(b2 Balance) {
	if b.hasPayins() {
		if b2.hasPayins() {
			log.Fatal("Cannot merge payins")
		}
	} else {
		b.AvgDate = b2.AvgDate
	}

	for _, cur := range balanceMaps() {
		b.Payins[cur].Value += b2.Payins[cur].Value
		b.Assets[cur].Value += b2.Assets[cur].Value
		b.Commissions[cur].Value += b2.Commissions[cur].Value
	}
}

func (b *Balance) CalcAllAssets(usd, eur float64) float64 {
	return b.Assets.CalcAll(usd, eur)
}

// =============================================================================

/*

 Balance consists of:

 USD
 + Assets:
    1. Cash balance
        1.1 Direct payins
        1.2 Exchanges
        1.3 Sold stocks
        1.4 - Bought stocks
        1.5 - Service commissions
        1.6 - Tax
        1.7 Dividends, coupons & repayments
    2. Open USD positions
 - Payins
    3. Directs payins
    4. Exchanges

RUB
 + Assets:
    1. Cash balance
        1.1 Direct payins
        1.3 Sold stocks & dollars
        1.4 - Bought stocks & dollars
        1.5 - Service commissions
        1.6 - Tax
        1.7 Dividends, coupons & repayments
    2. Open RUB positions
 - Payins:
    3. Direct payins
    5. - Exchanged money

*/

func (bal *Balance) AddOperation(op Operation, xchgrate func(curr_from, curr_to string, t time.Time) float64) {
	if op.IsTrading() || op.OperationType == "BrokerCommission" {
		// not accounted here

	} else if op.IsPayment() {
		// 1.7
		bal.Assets[op.Currency].Value += op.Payment
	} else if op.OperationType == "PayIn" {
		// 1.1
		bal.Assets[op.Currency].Value += op.Payment
		// 3
		bal.Payins[op.Currency].Value += op.Payment

		// add total payin
		payin := op.Payment * xchgrate(op.Currency, "RUB", op.DateParsed)
		if bal.Payins["all"].Value == 0 {
			bal.AvgDate = op.DateParsed
		} else {
			bal.AvgDate = aux.AdjustDate(bal.AvgDate, op.DateParsed, bal.Payins["all"].Value, payin)
		}
		bal.Payins["all"].Value += payin

	} else if op.OperationType == "ServiceCommission" {

		bal.Commissions[op.Currency].Value += op.Payment
		// add total
		bal.Commissions["all"].Value += op.Payment * xchgrate(op.Currency, "RUB", op.DateParsed)

		// 1.5
		bal.Assets[op.Currency].Value -= -op.Payment

	} else if op.OperationType == "Tax" {
		// 1.6
		bal.Assets[op.Currency].Value -= -op.Payment
	} else {
		log.Warnf("Unprocessed transaction 2 %v", op)
	}
}

func (bal *Balance) AddDeal(deal Deal, figi string) {
	if figi == FigiUSD {
		// Exchanges
		// 1.2
		bal.Assets["USD"].Value += float64(deal.Quantity)
		// 4
		bal.Payins["USD"].Value += float64(deal.Quantity)
		// 5
		bal.Payins["RUB"].Value -= deal.Value()
	}
	// 1.3, 1.4, 2
	bal.Assets[deal.Price.Currency].Value -= deal.Value() - deal.Commission
}

// =============================================================================

type SectionedBalance struct {
	Sections map[Section]*Balance
	Total    *Balance
}

func NewSectionedBalance() SectionedBalance {
	return SectionedBalance{
		Sections: make(map[Section]*Balance),
		Total:    NewBalance(),
	}
}

func (sb SectionedBalance) SectionBalance(section Section) *Balance {
	b := sb.Sections[section]
	if b == nil {
		b = NewBalance()
		sb.Sections[section] = b
	}
	return b
}

func (sb SectionedBalance) AddDeal(deal Deal, figi string, section Section) {
	sb.SectionBalance(section).AddDeal(deal, figi)
	sb.Total.AddDeal(deal, figi)
}

func (sb SectionedBalance) CalcAllAssets(usd, eur float64) {
	sb.Total.CalcAllAssets(usd, eur)
	for _, b := range sb.Sections {
		b.CalcAllAssets(usd, eur)
	}
}

const (
	TableStyle = "table"
)

func PrintBalanceHead(style string) {
	if style == TableStyle {
		fmt.Println("payins, assets, delta, bonds.rub, bonds.usd, stocks.rub, stocks.usd, pivotdate")
	}
}

func (b SectionedBalance) Print(prefix, style string) {
	p, a := b.Total.Payins["all"].Value, b.Total.Assets["all"].Value
	d := a - p

	sectionShare := func(section Section) float64 {
		if b.Sections != nil {
			if bal := b.Sections[section]; bal != nil {
				return 100 * bal.Assets["all"].Value / a
			}
		}
		return 0
	}

	bru, bus, sru, sus := math.Round(sectionShare(BondRub)*10)/10,
		math.Round(sectionShare(BondUsd)*10)/10,
		math.Round(sectionShare(StockRub)*10)/10,
		math.Round(sectionShare(StockUsd)*10)/10

	s := ""
	if style == TableStyle {
		if prefix != "" {
			s = prefix + ", "
		}
		s += fmt.Sprintf("%.0f, %.0f, %.0f, %.1f, %.1f, %.1f, %.1f, %s",
			p, a, d, bru, bus, sru, sus, b.Total.AvgDate.Format("2006/01/02"))

	} else {
		if prefix != "" {
			s = prefix + ": "
		}
		s += fmt.Sprintf("%7.0f -> %7.0f : %6.0f : bonds(R+U) %2.1f+%2.1f%% stocks %2.1f+%2.1f%% : pd %s",
			p, a, d, bru, bus, sru, sus, b.Total.AvgDate.Format("2006/01/02"))
	}
	fmt.Println(s)
}
