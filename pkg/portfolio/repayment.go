package portfolio

import (
	"time"
)

func (p *Portfolio) addStaticRepayments() {
	if pinfo, ok := p.positions["BBG00GW0RM55"]; ok {
		pinfo.AddRepayment(time.Date(2019, 12, 10, 7, 0, 0, 0, time.UTC),
			83)
		pinfo.AddRepayment(time.Date(2020, 3, 10, 7, 0, 0, 0, time.UTC),
			83)
	}
}

func (p *Portfolio) calculateRepayments() {
	amounts := make(map[string]int)

	for _, op := range p.data.ops {
		if op.Status != "Done" {
			continue
		}

		if op.IsTrading() {
			amounts[op.Figi] += op.Quantity()
			continue
		}

		if op.OperationType == "PartRepayment" {
			p.positions[op.Figi].AddRepayment(op.DateParsed, op.Payment/float64(amounts[op.Figi]))
			continue
		}
	}

	// Temporary fixup:
	// The problem is that the solution doesn't work for
	//  - amortized bonds
	//  - that were open at some t1 (point we want to know balance at)
	//  - but were all sold at some point t2
	//  - and there were more repayments from t2 till now
	// ..because there seems to be no way to get their partrepayment stats after the selling point
	// Maybe extrapolate the previous repayments?
	p.addStaticRepayments()

	// Now normalize the multipliers
	for _, pinfo := range p.positions {
		pinfo.RepaymentsNormalize()
	}
}
