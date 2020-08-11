package portfolio

import (
	"fmt"
	"math"
	"sort"

	"../schema"
)

func printBalanceHead(style string) {
	if style == schema.TableStyle {
		fmt.Println("payins, assets, delta, bonds.rub, bonds.usd, stocks.rub, stocks.usd, pivotdate")
	}
}

func printBalance(sections map[schema.Section]*schema.Balance, total *schema.Balance, prefix, style string) {
	p, a := total.Payins["all"].Value, total.Assets["all"].Value
	d := a - p

	sectionShare := func(section schema.Section) float64 {
		bal := sections[section]
		if bal == nil {
			return 0
		}

		return 100 * bal.Assets["all"].Value / a
	}

	bru, bus, sru, sus := math.Round(sectionShare(schema.BondRub)*10)/10,
		math.Round(sectionShare(schema.BondUsd)*10)/10,
		math.Round(sectionShare(schema.StockRub)*10)/10,
		math.Round(sectionShare(schema.StockUsd)*10)/10

	s := ""
	if style == schema.TableStyle {
		if prefix != "" {
			s = prefix + ", "
		}
		s += fmt.Sprintf("%.0f, %.0f, %.0f, %.1f, %.1f, %.1f, %.1f, %s",
			p, a, d, bru, bus, sru, sus, total.AvgDate.Format("2006/01/02"))

	} else {
		if prefix != "" {
			s = prefix + ": "
		}
		s += fmt.Sprintf("%7.0f -> %7.0f : %6.0f : bonds(R+U) %2.1f+%2.1f%% stocks %2.1f+%2.1f%% : pd %s",
			p, a, d, bru, bus, sru, sus, total.AvgDate.Format("2006/01/02"))
	}
	fmt.Println(s)
}

func (p *Portfolio) Print() {
	fmt.Println("== Totals ==")

	printBalance(p.sections, p.totals, "", "")

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
