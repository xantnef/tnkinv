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

func (p *Portfolio) makePortionYields(pinfo *schema.PositionInfo) {
	for _, po := range pinfo.Portions {
		var expense float64

		value := -po.Close.Value()

		for _, div := range pinfo.Dividends {
			if div.Date.Before(po.Buys[0].Date) {
				continue
			}
			if div.Date.After(po.Close.Date) {
				// TODO not quite right. Dividends come with delay
				continue
			}
			value += div.Value
		}

		for _, deal := range po.Buys {
			expense += deal.Value()
			expense += -deal.Commission
		}
		expense += -po.Close.Commission

		po.Yield = aux.Ratio2Perc(value / expense)
		po.YieldAnnual = aux.Ratio2Perc(aux.RatioAnnual(value/expense, po.Close.Date.Sub(po.AvgDate)))
		// compare with the market ETF
		po.YieldMarket = p.getMarketYield(pinfo.Ins, po, expense)

		po.Balance.Value = value - expense
		po.Balance.Currency = po.Close.Price.Currency
	}
}
