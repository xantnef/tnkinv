package portfolio

import (
	"fmt"
	"sort"

	"../schema"
)

func (p *Portfolio) Print() {
	fmt.Println("== Totals ==")

	p.totals.Print(p.sections, "", "")

	fmt.Printf(" alpha: %s\n", p.alphas)

	fmt.Println("== Current positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if pinfo.IsClosed() {
			return
		}
		fmt.Print("  " + pinfo.StringPretty() + "\n")
	})

	fmt.Println("== Closed positions ==")
	p.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if !pinfo.IsClosed() {
			return
		}
		fmt.Print("  " + pinfo.StringPretty() + "\n")
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
