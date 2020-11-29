package portfolio

import (
	"time"

	log "github.com/sirupsen/logrus"

	"../schema"
)

func (p *Portfolio) insByFigi(figi string) schema.Instrument {
	ins, ok := p.instruments[figi]
	if !ok {
		ins = p.client.RequestByFigi(figi)
		p.instruments[figi] = ins
	}
	log.Debug(ins)
	return ins
}

func (p *Portfolio) insByTicker(ticker string) schema.Instrument {
	for _, ins := range p.instruments {
		if ins.Ticker == ticker {
			return ins
		}
	}

	ins := p.client.RequestByTicker(ticker)
	p.instruments[ins.Figi] = ins
	return ins
}

func (p *Portfolio) tryGetTicker(figi string) string {
	if figi == "" {
		return ""
	}
	return p.insByFigi(figi).Ticker
}

func (p *Portfolio) benchPricef(ins schema.Instrument) schema.PriceAt {
	bench := ins.Benchmark()
	if bench == "" {
		return nil
	}

	bins := p.insByTicker(bench)

	return func(t time.Time) float64 {
		return p.cc.GetInCurrency(bins, ins.Currency, t)
	}
}
