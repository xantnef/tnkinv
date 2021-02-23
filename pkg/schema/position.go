package schema

import (
	"fmt"
	"sort"
	"time"

	"../aux"
	log "github.com/sirupsen/logrus"
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
	OpenDeal     Deal

	// TODO commissions are counted here but not included in portion balances and yields
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
		}
		pinfo.Portions = append(pinfo.Portions, po)
	}

	pinfo.OpenQuantity += deal.Quantity

	if pinfo.OpenQuantity < 0 {
		log.Fatalf("negative balance? %v", pinfo)
	}

	if len(po.Buys) > 0 {
		last := po.Buys[len(po.Buys)-1]

		if deal.Date.Before(last.Date.Add(12 * time.Hour)) {
			po.Buys = po.Buys[:len(po.Buys)-1]

			// merge deals
			sval := deal.Value() + last.Value()
			deal.Quantity += last.Quantity
			deal.Accrued += last.Accrued
			deal.Commission += last.Commission

			deal.Price.Value = (sval - deal.Accrued) / float64(deal.Quantity)
		}
	}

	if pinfo.OpenQuantity > 0 {
		po.Buys = append(po.Buys, deal)
	} else {
		// complete sell
		po.finalize(deal, true)
	}
}

func (pinfo *PositionInfo) AddOperation(op Operation) (Deal, bool) {
	log.Debugf("%v", op)

	if op.Status != "Done" {
		return Deal{}, false
	}

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

// =============================================================================

func (pinfo *PositionInfo) MakeOpenDeal(date time.Time, pricef func() float64) (deal Deal, ok bool) {
	po := pinfo.openPortion()
	if po == nil {
		return deal, false
	}

	deal = Deal{
		Date:     date,
		Price:    NewCValue(pricef(), po.Balance.Currency),
		Quantity: -pinfo.OpenQuantity,
	}

	po.finalize(deal, false)
	pinfo.OpenDeal = deal

	return deal, true
}

func (pinfo *PositionInfo) Finalize(benchPricef PriceAt) {
	for _, po := range pinfo.Portions {
		var xirr aux.XirrCtx
		value := -po.Close.Value()
		expense := -po.Close.Commission

		for _, div := range pinfo.Dividends {
			if div.Date.Before(po.Buys[0].Date) {
				continue
			}
			if div.Date.After(po.Close.Date) {
				// TODO not quite right. Dividends come with delay
				continue
			}
			value += div.Value
			xirr.AddPayment(-div.Value, div.Date)
		}

		for _, deal := range po.Buys {
			if deal.IsBuy() {
				expense += deal.Expense()
			} else {
				value += deal.Profit()
			}

			xirr.AddPayment(deal.Expense(), deal.Date)
		}

		// there are fictive deals with 0 quantity
		if expense != 0 {
			po.Yield = aux.Ratio2Perc(value / expense)
			po.YieldAnnual = xirr.Ratio(po.Close.Profit(), po.Close.Date) * 100
			// compare with the market ETF
			if benchPricef != nil {
				po.YieldMarket = aux.Ratio2Perc(po.benchValue(benchPricef) / expense)
			}
		}

		po.Balance.Value = value - expense
		po.Balance.Currency = po.Close.Price.Currency
	}
}

func (pinfo PositionInfo) Alpha() CValue {
	alpha := NewCValue(0, pinfo.Ins.Currency)

	if pinfo.Ins.Type == InsTypeEtf {
		return alpha
	}

	for _, po := range pinfo.Portions {
		alpha.Value += po.Alpha().Value
	}

	return alpha
}

// =============================================================================

// TODO these dont look good at all

func (pinfo *PositionInfo) AddRepayment(t time.Time, value float64) {
	pinfo.Repayments = append(pinfo.Repayments,
		&RepaymentPoint{
			Time: t,
			Mult: 1,
		})

	for _, rep := range pinfo.Repayments {
		rep.Mult += value / float64(pinfo.Ins.FaceValue)
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
