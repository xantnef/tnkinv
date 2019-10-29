package myclient

import (
	"fmt"
	"sort"

	"../schema"
)

func (c *myClient) currExchangeDiff(currency string) (diff float64) {
	uspos := c.positions[schema.FigiUSD]
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

func (c *myClient) printPortfolio() {
	fmt.Println("== Totals ==")

	fmt.Println("  Payins:")
	for _, cv := range c.totals.payins {
		if cv.Value == 0 {
			continue
		}
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Commissions:")
	fmt.Printf("    %v\n", c.totals.commission)

	fmt.Println("  Assets:")
	for _, cv := range c.totals.assets {
		if cv.Value == 0 {
			continue
		}
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Balance:")
	for currency, cv := range c.totals.payins {
		bal := schema.NewCValue(
			c.totals.assets[currency].Value-cv.Value,
			currency)
		bal.Value += c.currExchangeDiff(currency)
		if bal.Value == 0 {
			continue
		}

		fmt.Printf("    %v\n", bal)
	}

	fmt.Println("== Current positions ==")
	c.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if pinfo.IsClosed {
			return
		}
		fmt.Print("  " + pinfo.String() + "\n")
	})

	fmt.Println("== Closed positions ==")
	c.forSortedPositions(func(pinfo *schema.PositionInfo) {
		if !pinfo.IsClosed {
			return
		}
		fmt.Print("  " + pinfo.String() + "\n")
	})
}

func (c *myClient) forSortedPositions(cb func(pinfo *schema.PositionInfo)) {
	if len(c.figisSorted) == 0 {
		var figis []string
		for figi := range c.positions {
			figis = append(figis, figi)
		}

		sort.Slice(figis, func(i, j int) bool {
			p1 := c.positions[figis[i]]
			p2 := c.positions[figis[j]]

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
		c.figisSorted = figis
	}

	for _, figi := range c.figisSorted {
		cb(c.positions[figi])
	}
}
