package portfolio

import (
	"fmt"
	"time"

	"../aux"
	"../candles"
	"../client"
	"../schema"
)

func GetPrices(c *client.MyClient, tickers []string, start, end time.Time) {
	cc := candles.NewCandleCache(c)

	for _, ticker := range tickers {
		ins := c.RequestByTicker(ticker)
		p1 := cc.Get(ins.Figi, start)
		p2 := cc.Get(ins.Figi, end)

		s := fmt.Sprintf("%s: %.2f -> %.2f (%.1f%% %s; %.1f%% annual)",
			ticker, p1, p2, aux.Ratio2Perc(p2/p1), ins.Currency,
			aux.Ratio2Perc(aux.RatioAnnual(p2/p1, end.Sub(start))))

		if ins.Currency != "RUB" {
			p1 *= cc.Get(schema.FigiUSD, start)
			p2 *= cc.Get(schema.FigiUSD, end)

			s += fmt.Sprintf(" (%.1f%% RUB; %.1f%% annual)",
				aux.Ratio2Perc(p2/p1),
				aux.Ratio2Perc(aux.RatioAnnual(p2/p1, end.Sub(start))))

		} else if section, ok := schema.GetEtfSection(ticker); ok {

			if cur := schema.GetSectionCurrency(section); cur == "USD" {
				p1 /= cc.Get(schema.FigiUSD, start)
				p2 /= cc.Get(schema.FigiUSD, end)
				s += fmt.Sprintf(" (%.1f%% USD; %.1f%% annual)",
					aux.Ratio2Perc(p2/p1),
					aux.Ratio2Perc(aux.RatioAnnual(p2/p1, end.Sub(start))))
			}
		}

		fmt.Println(s)
	}
}
