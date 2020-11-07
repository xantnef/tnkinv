package schema

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func balanceMaps() []string {
	lst := []string{"all"}
	return append(lst, CurrenciesOrdered[:]...)
}

type CurMap map[string]*CValue

type Balance struct {
	Commissions, Payins, Assets CurMap
	AvgDate                     time.Time
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

	for _, cur := range balanceMaps() {
		if b.Payins[cur].Value != 0 {
			actualCurrencies = append(actualCurrencies, cur)
		}
	}

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
