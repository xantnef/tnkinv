package schema

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"../aux"
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

	Deals      []Deal
	Dividends  []Dividend
	Portions   []*Portion
	Repayments []*RepaymentPoint

	OpenQuantity int
	OpenSpent    float64

	OpenDeal Deal

	AccumulatedIncome CValue
}

func (pinfo PositionInfo) IsClosed() bool {
	return pinfo.OpenDeal.Quantity == 0
}

func (pinfo PositionInfo) StringPretty() string {
	s := fmt.Sprintf("%s:", pinfo.Ins.Name)

	if !pinfo.IsClosed() {
		od := pinfo.OpenDeal
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

func (pinfo PositionInfo) openPortion() *Portion {
	if len(pinfo.Portions) == 0 {
		return nil
	}

	po := pinfo.Portions[len(pinfo.Portions)-1]
	if po.IsClosed {
		return nil
	}

	return po
}

func (pinfo *PositionInfo) addDeal(deal Deal) {
	pinfo.Deals = append(pinfo.Deals, deal)

	po := pinfo.openPortion()
	if po == nil {
		po = &Portion{
			Balance: NewCValue(0, deal.Price.Currency),
			AvgDate: deal.Date,
		}
		pinfo.Portions = append(pinfo.Portions, po)
	}

	pinfo.OpenQuantity += deal.Quantity
	pinfo.OpenSpent += deal.Value()

	if deal.Quantity > 0 { // buy
		po.CheckNoSplitSells(pinfo.Ins.Ticker)

		po.AvgDate = aux.AdjustDate(po.AvgDate, deal.Date, pinfo.OpenSpent, deal.Value())

		//po.AvgPrice.Value = deal.Price.Value*mult + po.AvgPrice.Value*(1-mult)
		po.Buys = append(po.Buys, deal)

	} else { // sell

		if pinfo.OpenQuantity < 0 {
			log.Fatalf("negative balance? %v", pinfo)
		}

		if pinfo.OpenQuantity > 0 {
			// How to better handle it?
			// 1. split sells. try and merge. those wont split between days,
			//    so wont cross portion period boundaries
			//
			// 2. true partial sells. options?
			//    2.1 sell all, buy some back
			//    2.2 ?
			//
			po.SplitSells = append(po.SplitSells, deal)

		} else {
			if len(po.SplitSells) > 0 {
				// new "superdeal"
				sdeal := deal
				sval := sdeal.Value()

				for _, psell := range po.SplitSells {
					sdeal.Quantity += psell.Quantity
					sdeal.Accrued += psell.Accrued
					sdeal.Commission += psell.Commission
					sval += psell.Value()
				}

				sdeal.Price.Value = (sval - sdeal.Accrued) / float64(sdeal.Quantity)
				deal = sdeal
			}

			pinfo.OpenSpent = 0

			// complete sell
			po.Close = deal
			po.IsClosed = true
		}
	}
}

func (pinfo *PositionInfo) AddOperation(op Operation) (Deal, bool) {
	log.Debugf("%s", op)

	if op.IsTrading() {
		deal := Deal{
			Date:       op.DateParsed,
			Price:      NewCValue(op.Price, op.Currency),
			Quantity:   op.Quantity(),
			Commission: op.Commission.Value,
		}

		// op.Payment is negative for Buy
		// deal.Quantity is positive for Buy
		// deal.Price is always positive
		// Commission is not included in Payment
		deal.Accrued = -op.Payment - deal.Price.Value*float64(deal.Quantity)

		pinfo.addDeal(deal)

		return deal, true

	} else if op.OperationType == "BrokerCommission" {
		// negative
		pinfo.AccumulatedIncome.Value += op.Payment

	} else if op.IsPayment() {
		// income - positive, taxes - negative
		pinfo.AccumulatedIncome.Value += op.Payment
		pinfo.Dividends = append(pinfo.Dividends,
			Dividend{
				Date:  op.DateParsed,
				Value: op.Payment,
			})
	} else if op.OperationType == "Tax" {
		// negative
		pinfo.AccumulatedIncome.Value += op.Payment
	} else {
		log.Warnf("Unprocessed transaction %v", op)
	}

	return Deal{}, false
}

func (pinfo *PositionInfo) MakeOpenDeal(date time.Time, pricef func() float64) (deal Deal, ok bool) {
	po := pinfo.openPortion()
	if po == nil {
		return deal, false
	}

	po.CheckNoSplitSells(pinfo.Ins.Ticker)

	deal = Deal{
		Date:     date,
		Price:    NewCValue(pricef(), po.Balance.Currency),
		Quantity: -pinfo.OpenQuantity,
	}

	po.Close = deal
	pinfo.OpenDeal = deal

	return deal, true
}

// =============================================================================

// TODO these dont look good at all

func (pinfo *PositionInfo) AddRepayment(t time.Time, value float64) {
	pinfo.Repayments = append(pinfo.Repayments,
		&RepaymentPoint{
			Time: t,
		})

	for _, rep := range pinfo.Repayments {
		rep.Mult += value
	}
}

func (pinfo *PositionInfo) RepaymentsNormalize() {
	for _, rep := range pinfo.Repayments {
		log.Debugf("repayment %s at %s: %f/%d", pinfo.Ins.Name, rep.Time, rep.Mult, pinfo.Ins.FaceValue)
		rep.Mult = (rep.Mult + float64(pinfo.Ins.FaceValue)) / float64(pinfo.Ins.FaceValue)
	}
}

func (pinfo PositionInfo) RepaymentMultiplier(t time.Time) float64 {
	idx := sort.Search(len(pinfo.Repayments), func(i int) bool {
		return pinfo.Repayments[i].Time.After(t)
	})

	if idx < len(pinfo.Repayments) {
		return pinfo.Repayments[idx].Mult
	}

	return 1
}
