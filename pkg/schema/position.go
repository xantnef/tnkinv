package schema

import (
	"fmt"
	"time"
)

type Dividend struct {
	Date  time.Time
	Value float64
}

type RepaymentPoint struct {
	Time time.Time
	Mult float64
}

type PositionInfo struct {
	Ins Instrument

	Deals      []*Deal
	Dividends  []*Dividend
	Portions   []*Portion
	Repayments []*RepaymentPoint

	OpenQuantity int // TODO
	OpenSpent    float64
	OpenDeal     *Deal

	AccumulatedIncome CValue // TODO
}

func (pinfo *PositionInfo) IsClosed() bool {
	return pinfo.OpenDeal == nil
}

func (pinfo *PositionInfo) StringPretty() string {
	s := fmt.Sprintf("%s:", pinfo.Ins.Name)

	od := pinfo.OpenDeal
	if od != nil {
		s += fmt.Sprintf(" %s (%.2f x %d) +acc %v",
			od.CValue().Mult(-1.0),
			od.Price.Value, -od.Quantity, pinfo.AccumulatedIncome)
	}

	s += "\n" +
		"    deals:\n"

	for _, deal := range pinfo.Deals {
		s += "      " + deal.String() + "\n"
	}

	s += "    position stats:\n"

	for _, po := range pinfo.Portions {
		s += "      " + po.String() + "\n"
	}

	return s
}
