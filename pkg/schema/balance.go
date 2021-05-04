package schema

import (
	"fmt"
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

	xirr aux.XirrCtx
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
	copy.xirr = b.xirr

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
		b.xirr = b2.xirr
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
		bal.xirr.AddPayment(payin, op.DateParsed)
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

	log.Debugf("Added %s to %s", op.Figi, bal.Assets)
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

func (sb SectionedBalance) sectionShare(section Section) float64 {
	if sb.Sections != nil && sb.Total != nil {
		if bal := sb.Sections[section]; bal != nil {
			a := sb.Total.Assets["all"].Value
			return 100 * bal.Assets["all"].Value / a
		}
	}
	return 0
}

const (
	TableStyle = "table"
)

func PrintBalanceHead(style string) {
	if style == TableStyle {
		fmt.Println("payins, assets, delta, bonds.rub, bonds.usd, stocks.rub, stocks.usd, pivotdate")
	}
}

func (b SectionedBalance) Print(t time.Time, prefix, style string) {
	p, a := b.Total.Payins["all"].Value, b.Total.Assets["all"].Value
	d := a - p

	s := ""
	if style == TableStyle {
		if prefix != "" {
			s = prefix + ", "
		}
		s += fmt.Sprintf("%.0f, %.0f, %.0f, %.1f, %.1f, %.1f, %.1f",
			p, a, d,
			b.sectionShare(BondRub),
			b.sectionShare(BondUsd),
			b.sectionShare(StockRub),
			b.sectionShare(StockUsd))
	} else {
		if prefix != "" {
			s = prefix + ": "
		}
		s += fmt.Sprintf("%7.0f -> %7.0f : %6.0f (%5.1f%%, annual %5.1f%%) "+
			"bonds(R+U): %5.1f + %5.1f%%; stocks: %5.1f+%5.1f%%",
			p, a, d,
			aux.Ratio2Perc(a/p), b.Total.xirr.Ratio(a, t)*100,
			b.sectionShare(BondRub),
			b.sectionShare(BondUsd),
			b.sectionShare(StockRub),
			b.sectionShare(StockUsd))
	}
	fmt.Println(s)
}
