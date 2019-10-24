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
	var figis []string
	for figi := range c.positions {
		figis = append(figis, figi)
	}
	sort.Strings(figis)
	for _, figi := range figis {
		cb(c.positions[figi])
	}
}
