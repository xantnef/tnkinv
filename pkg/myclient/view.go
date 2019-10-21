package myclient

import (
	"fmt"

	"../schema"
)

func (c *myClient) printPortfolio() {
	fmt.Println("== Totals ==")

	fmt.Println("  Payins:")
	for _, cv := range c.totals.payins {
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Commissions:")
	fmt.Printf("    %v\n", c.totals.commission)

	fmt.Println("  Assets:")
	for _, cv := range c.totals.assets {
		fmt.Printf("    %v\n", *cv)
	}

	fmt.Println("  Balance:")
	for currency, cv := range c.totals.payins {
		bal := schema.NewCValue(
			c.totals.assets[currency].Value-cv.Value,
			currency)
		fmt.Printf("    %v\n", bal)
	}

	fmt.Println("== Current positions ==")
	for _, pinfo := range c.positions {
		if pinfo.IsClosed {
			continue
		}
		fmt.Print("  " + pinfo.String())
	}

	fmt.Println("== Closed positions ==")
	for _, pinfo := range c.positions {
		if !pinfo.IsClosed {
			continue
		}
		fmt.Print("  " + pinfo.String())
	}
}
