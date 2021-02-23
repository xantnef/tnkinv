package portfolio

import (
	"fmt"
	"time"

	"../aux"
	"../candles"
	"../client"
	"../schema"
)

type price struct {
	time  time.Time
	price float64
}

type history struct {
	ins    schema.Instrument
	prices []price
}

func printTotal(cc *candles.CandleCache, ins schema.Instrument, start, end price) {
	s := fmt.Sprintf("%s: %.2f -> %.2f (%.1f%% %s; %.1f%% annual)",
		ins.Ticker, start.price, end.price, aux.Ratio2Perc(end.price/start.price), ins.Currency,
		aux.Ratio2Perc(aux.RatioAnnual(end.price/start.price, end.time.Sub(start.time))))

	if ins.Currency != "RUB" {
		start.price = cc.GetInCurrency(ins, "RUB", start.time)
		end.price = cc.GetInCurrency(ins, "RUB", end.time)

		s += fmt.Sprintf(" (%.1f%% RUB; %.1f%% annual)",
			aux.Ratio2Perc(end.price/start.price),
			aux.Ratio2Perc(aux.RatioAnnual(end.price/start.price, end.time.Sub(start.time))))

	} else if section, ok := schema.GetEtfSection(ins.Ticker); ok {
		if curr := section.Currency(); curr != ins.Currency {
			start.price = cc.GetInCurrency(ins, curr, start.time)
			end.price = cc.GetInCurrency(ins, curr, end.time)
			s += fmt.Sprintf(" (%.1f%% USD; %.1f%% annual)",
				aux.Ratio2Perc(end.price/start.price),
				aux.Ratio2Perc(aux.RatioAnnual(end.price/start.price, end.time.Sub(start.time))))
		}
	}

	fmt.Println(s)
}

func printHuman(cc *candles.CandleCache, hs []history) {
	s := fmt.Sprintf("%-10s ", "date")

	for _, h := range hs {
		s += fmt.Sprintf("%6s ", h.ins.Ticker)
	}
	fmt.Println(s)

	for i, p := range hs[0].prices {
		s := p.time.Format("2006/01/02 ")
		for _, h := range hs {
			s += fmt.Sprintf("%6.1f ", aux.Ratio2Perc(h.prices[i].price/h.prices[0].price))
		}
		fmt.Println(s)
	}

	fmt.Println("--")

	for _, h := range hs {
		printTotal(cc, h.ins, h.prices[0], h.prices[len(h.prices)-1])
	}
}

func printTable(hs []history) {
	s := "date"
	for _, h := range hs {
		s += ", " + h.ins.Ticker
	}
	fmt.Println(s)

	for i, p := range hs[0].prices {
		s := p.time.Format("2006/01/02")
		for _, h := range hs {
			s += fmt.Sprintf(", %.1f", aux.Ratio2Perc(h.prices[i].price/h.prices[0].price))
		}
		fmt.Println(s)
	}
}

func GetPrices(c *client.MyClient, tickers []string, start, end time.Time, period, format string) {
	hs := make([]history, len(tickers))
	times := []time.Time{}
	curr := ""

	cc := candles.NewCandleCache(c)

	if period == "" {
		times = []time.Time{start, end}
	} else {
		cc = cc.WithPeriod(start, period)
		times = cc.ListTimes()
	}

	for i, ticker := range tickers {
		hs[i] = history{
			ins:    c.RequestByTicker(ticker),
			prices: make([]price, len(times)),
		}
		if curr == "" {
			curr = hs[i].ins.Currency
		} else if curr != hs[i].ins.Currency {
			curr = "RUB"
		}
	}

	for i := range hs {
		h := &hs[i]
		for i, t := range times {
			h.prices[i] = price{
				time:  t,
				price: cc.GetInCurrency(h.ins, curr, t),
			}
		}
	}

	if format == "human" {
		printHuman(cc, hs)
	} else {
		printTable(hs)
	}
}
