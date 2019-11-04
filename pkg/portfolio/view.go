package portfolio

import (
	"fmt"
	"sort"

	"../schema"
)

func (p *Portfolio) currExchangeDiff(currency string) (diff float64) {
	uspos := p.positions[schema.FigiUSD]
	if uspos == nil {
		return
	}

	for _, deal := range uspos.Deals {
		if currency == "RUB" {
			diff += deal.Price.Value * float64(deal.Quantity)
		}
		if currency == "USD" {
			diff -= float64(deal.Quantity)
		}
	}
	return
}

func (p *Portfolio) Print() {
	fmt.Println("== Totals ==")

	fmt.Println("  Payins:")
	for _, cv := range p.totals.payins {
		if cv.Value == 0 {
			continue
		}
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Commissions:")
	fmt.Printf("    %v\n", p.totals.commission)

	fmt.Println("  Assets:")
	for _, cv := range p.totals.assets {
		if cv.Value == 0 {
			continue
		}
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Balance:")
	for currency, cv := range p.totals.payins {
		bal := schema.NewCValue(
			p.totals.assets[currency].Value-cv.Value,
			currency)
		bal.Value += p.currExchangeDiff(currency)
		if bal.Value == 0 {
			continue
		}

		fmt.Printf("    %v\n", bal)
	}

	fmt.Println("== Current positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if pinfo.IsClosed {
			return
		}
		fmt.Print("  " + pinfo.String() + "\n")
	})

	fmt.Println("== Closed positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if !pinfo.IsClosed {
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
			p1 := p.positions[figis[i]]
			p2 := p.positions[figis[j]]

			if p1.IsClosed && !p2.IsClosed {
				return false
			}
			if !p1.IsClosed && p2.IsClosed {
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
