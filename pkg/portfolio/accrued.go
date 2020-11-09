package portfolio

import (
	"time"

	log "github.com/sirupsen/logrus"

	"../schema"
)

func (p *Portfolio) collectAccrued() {
	p.config.enableAccrued = true

	for _, acc := range p.accs {
		pfResp := p.client.RequestPortfolio(acc)
		for _, pos := range pfResp.Payload.Positions {
			p.accrued[pos.Figi] = pos.AveragePositionPrice.Value - pos.AveragePositionPriceNoNkd.Value
		}
	}
}

func (p *Portfolio) getAccrued(pinfo *schema.PositionInfo, date time.Time) float64 {
	if !p.config.enableAccrued {
		return 0
	}
	if pinfo.Ins.Type != schema.InsTypeBond {
		return 0
	}

	// Accrued value cannot be fetched for date != Now
	if time.Now().Sub(date).Hours() > 24 {
		log.Warnf("missing accrued value for %s, balance is inaccurate", pinfo.Ins.Figi)
		return 0
	}

	accrued, ok := p.accrued[pinfo.Ins.Figi]
	if !ok {
		log.Warnf("missing accrued value for %s, balance is inaccurate", pinfo.Ins.Figi)
		return 0
	}
	return accrued
}
