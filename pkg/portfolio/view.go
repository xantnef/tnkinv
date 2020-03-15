package portfolio

import (
	"fmt"
	"sort"

	"../schema"
)

func (p *Portfolio) Print() {
	fmt.Println("== Totals ==")

	p.totals.Foreach(func(nm string, m schema.CurMap) {
		fmt.Printf("  %s:\n", nm)
		for _, currency := range schema.CurrenciesOrdered {
			cv := m[currency]
			if cv.Value == 0 {
				continue
			}
			fmt.Printf("    %v\n", *cv)
		}
	})

	fmt.Println("  Balance:")
	for _, currency := range schema.CurrenciesOrdered {
		bal := p.totals.Get(currency)
		if bal.Value == 0 {
			continue
		}
		fmt.Printf("    %v\n", bal)
	}

	fmt.Println("  Distribution:")
	for _, currency := range schema.CurrenciesOrdered {
		if p.totals.Assets[currency].Value == 0 {
			continue
		}

		fmt.Printf("    %s:\n", currency) // TODO currency percentage

		types := []*schema.Balance{p.funds, p.stocks, p.bonds, p.cash}
		names := []string{schema.InsTypeEtf, schema.InsTypeStock, schema.InsTypeBond, "Cash"}

		for i, t := range types {
			if t == nil {
				continue
			}
			fmt.Printf("      %6s: %v (%.0f%%)\n", names[i],
				t.Assets[currency], 100*t.Assets[currency].Value/p.totals.Assets[currency].Value)
		}
	}

	fmt.Println("== Current positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if pinfo.IsClosed() {
			return
		}
		fmt.Print("  " + pinfo.String() + "\n")
	})

	fmt.Println("== Closed positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if !pinfo.IsClosed() {
			return
		}
		fmt.Print("  " + pinfo.String() + "\n")
	})
}

func (p *Portfolio) forSortedPositions(cb func(pinfo *schema.PositionInfo)) {
	if len(p.figisSorted) == 0 {
		var figis []string
		for figi := range p.positions {
			figis = append(figis, figi)
		}

		sort.Slice(figis, func(i, j int) bool {

			if figis[i] == schema.FigiUSD {
				return true
			}
			if figis[j] == schema.FigiUSD {
				return false
			}

			p1 := p.positions[figis[i]]
			p2 := p.positions[figis[j]]

			if p1.IsClosed() && !p2.IsClosed() {
				return false
			}
			if !p1.IsClosed() && p2.IsClosed() {
				return true
			}

			t1 := p1.Portions[len(p1.Portions)-1].Buys[0].Date
			t2 := p2.Portions[len(p2.Portions)-1].Buys[0].Date

			return t1.Before(t2)
		})
		p.figisSorted = figis
	}

	for _, figi := range p.figisSorted {
		cb(p.positions[figi])
	}
}
