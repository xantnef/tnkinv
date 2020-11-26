package portfolio

import (
	"../aux"
	"../schema"
)

func (p *Portfolio) getMarketYield(ins schema.Instrument, po *schema.Portion, expense float64) float64 {
	bench := ins.Benchmark()
	if bench == "" {
		return 0
	}

	bins := p.insByTicker(bench)

	var quantity float64
	for _, deal := range po.Buys {
		quantity += deal.Value() / p.cc.GetInCurrency(bins, ins.Currency, deal.Date)
	}

	value := quantity * p.cc.GetInCurrency(bins, ins.Currency, po.Close.Date)

	return aux.Ratio2Perc(value / expense)
}

func (p *Portfolio) makePortionYields(pinfo *schema.PositionInfo) schema.CValue {
	alpha := schema.NewCValue(0, pinfo.Ins.Currency)

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
			po.YieldMarket = p.getMarketYield(pinfo.Ins, po, expense)
		}

		po.Balance.Value = value - expense
		po.Balance.Currency = po.Close.Price.Currency

		alpha.Value += po.Alpha().Value
	}

	return alpha
}
