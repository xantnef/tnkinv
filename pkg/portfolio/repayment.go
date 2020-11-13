package portfolio

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (p *Portfolio) addStaticRepayments() {
	if pinfo, ok := p.positions["BBG00GW0RM55"]; ok {
		dates := []string{"2019/12/10", "2020/03/10"}
		for _, date := range dates {
			t, err := time.Parse("2006/01/02", date)
			if err != nil {
				log.Fatal(err)
			}
			pinfo.AddRepayment(t, 83)
		}
	}
}

// gotta calc them repayments first, to be able to get correct prices
// when calculating balances
func (p *Portfolio) preprocessOperations() {
	amounts := make(map[string]int)

	for _, op := range p.data.ops {
		if op.Status != "Done" {
			continue
		}

		if op.Figi == "" {
			continue
		}

		if op.IsTrading() {
			amounts[op.Figi] += op.Quantity()

		} else if op.OperationType == "PartRepayment" {
			pinfo := p.addPosition(op)
			pinfo.AddRepayment(op.DateParsed, op.Payment/float64(amounts[op.Figi]))
		}
	}

	// Temporary fixup:
	// The problem is that the repayment multiplier don't work for
	//  - amortized bonds
	//  - that were open at some t1 (point we want to know balance at)
	//  - but were all sold at some point t2
	//  - and there were more repayments from t2 till now
	// ..because there seems to be no way to get their partrepayment stats after the selling point
	// Maybe extrapolate the previous repayments?
	p.addStaticRepayments()
}
