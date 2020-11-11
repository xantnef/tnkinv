package schema

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Portion struct {
	Buys  []Deal
	Close Deal

	IsClosed bool

	AvgDate time.Time

	Balance     CValue
	Yield       float64
	YieldAnnual float64
	YieldMarket float64

	// aux and temporary
	SplitSells []Deal
}

func (po *Portion) CheckNoSplitSells(ticker string) {
	if len(po.SplitSells) > 0 {
		// those split sells were indeed partial
		log.Errorf("%s: partial sells are not handled nicely yet %s",
			ticker, po.SplitSells)
		po.SplitSells = []Deal{}
	}
}

func (po Portion) Alpha() CValue {
	if po.YieldMarket == 0 {
		return NewCValue(0, "RUB")
	}
	return po.Balance.Mult(1 - po.YieldMarket/po.Yield)
}

func (po Portion) String() string {
	date := "----/--/--"
	if po.IsClosed {
		date = po.Close.Date.Format("2006/01/02")
	}

	benchString := func(po Portion) string {
		if po.YieldMarket == 0 { // TODO not right, might have been real 0
			return ""
		}
		return fmt.Sprintf(", market %.1f%%, alpha %s", po.YieldMarket, po.Alpha())
	}

	return fmt.Sprintf(
		"%s: %s (%.1f%%, annual %.1f%%%s)",
		date, po.Balance, po.Yield, po.YieldAnnual, benchString(po))
}
