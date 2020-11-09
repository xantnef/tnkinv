package portfolio

import (
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
