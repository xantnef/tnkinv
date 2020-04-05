package portfolio

import (
	"fmt"
	"time"

	"../candles"
	"../client"
	"../schema"
)

func GetPrices(c *client.MyClient, tickers []string, start, end time.Time) {
	cc := candles.NewCandleCache(c, start, "day")

	for _, ticker := range tickers {
		ins := c.RequestByTicker(ticker)
		p1 := cc.GetOnDay(ins.Figi, start)
		p2 := cc.GetOnDay(ins.Figi, end)

		s := fmt.Sprintf("%s: %.2f -> %.2f (%.1f%% %s)",
			ticker, p1, p2, (p2/p1-1)*100, ins.Currency)

		if ins.Currency != "RUB" {
			p1 *= cc.GetOnDay(schema.FigiUSD, start)
			p2 *= cc.GetOnDay(schema.FigiUSD, end)

			s += fmt.Sprintf(" (%.1f%% RUB)", (p2/p1-1)*100)
		}

		fmt.Println(s)
	}
}
