package portfolio

import (
	"log"
	"math"
	"sort"
	"time"

	"../client"
	"../schema"
)

type Portfolio struct {
	client *client.MyClient

	tickers   map[string]string
	positions map[string]*schema.PositionInfo

	figisSorted []string

	totals *schema.Balance
}

func NewPortfolio(c *client.MyClient) *Portfolio {
	return &Portfolio{
		client:    c,
		tickers:   make(map[string]string),
		positions: make(map[string]*schema.PositionInfo),
		totals:    schema.NewBalance(),
	}
}

func (p *Portfolio) currExchangeDiff(currency string) (diff float64) {
	uspos := p.positions[schema.FigiUSD]
	if uspos == nil {
		return
	}

	for _, deal := range uspos.Deals {
		if currency == "RUB" {
			diff += deal.Value()
		}
		if currency == "USD" {
			diff -= float64(deal.Quantity)
		}
	}
	return
}

func (p *Portfolio) GetBalance(currency string) schema.CValue {
	bal := schema.NewCValue(
		p.totals.Assets[currency].Value-p.totals.Payins[currency].Value,
		currency)

	bal.Value += p.currExchangeDiff(currency)

	return bal
}

func (p *Portfolio) getTicker(figi string) string {
	ticker, ok := p.tickers[figi]
	if !ok {
		ticker = p.client.RequestTicker(figi)
		p.tickers[figi] = ticker
	}
	return ticker
}

func (p *Portfolio) Collect() error {
	c := p.client

	pfResp := c.RequestPortfolio()

	for _, pos := range pfResp.Payload.Positions {
		currency := pos.ExpectedYield.Currency
		p.tickers[pos.Figi] = pos.Ticker

		pinfo := &schema.PositionInfo{
			Figi:         pos.Figi,
			Ticker:       pos.Ticker,
			CurrentPrice: schema.NewCValue(c.RequestCurrentPrice(pos.Figi), currency),
			Quantity:     pos.Balance,

			AccumulatedIncome: schema.NewCValue(0, currency),
		}

		p.positions[pos.Figi] = pinfo

		if pos.Figi == schema.FigiUSD {
			p.totals.Assets["USD"].Value += pos.Balance
		} else {
			p.totals.Assets[currency].Value += pinfo.CurrentPrice.Value * pinfo.Quantity
		}
	}

	opsResp := c.RequestOperations()

	//log.Print("== Transaction log ==")

	for _, op := range opsResp.Payload.Operations {
		date, err := time.Parse(time.RFC3339, op.Date)
		if err != nil {
			log.Fatal("Failed to parse time: %v", err)
		}

		/* log.Printf("at %s %s some %s",
		date.String(), op.OperationType+"-ed",
		p.tickers[op.Figi])*/

		if op.Status != "Done" {
			// cancelled declined etc
			continue
		}

		if op.Figi == "" {
			if op.OperationType == "PayIn" {
				p.totals.Payins[op.Currency].Value += op.Payment
			} else if op.OperationType == "ServiceCommission" {
				p.totals.Commissions[op.Currency].Value += op.Payment
			}
			continue
		}

		pinfo := p.positions[op.Figi]
		if pinfo == nil {
			pinfo = &schema.PositionInfo{
				Figi:     op.Figi,
				Ticker:   p.getTicker(op.Figi),
				IsClosed: true,
			}
			p.positions[op.Figi] = pinfo
		}

		if op.OperationType == "Buy" || op.OperationType == "BuyCard" ||
			op.OperationType == "Sell" {
			deal := &schema.Deal{
				Date:       date,
				Price:      schema.NewCValue(op.Price, op.Currency),
				Quantity:   int(op.Quantity),
				Commission: op.Commission.Value,
			}
			if op.OperationType == "Sell" {
				deal.Quantity = -deal.Quantity
			}

			pinfo.Deals = append(pinfo.Deals, deal)

		} else if op.OperationType == "BrokerCommission" {
			// negative
			pinfo.AccumulatedIncome.Value += op.Payment
		} else if op.OperationType == "Dividend" || op.OperationType == "TaxDividend" {
			// positive, negative
			pinfo.AccumulatedIncome.Value += op.Payment
			pinfo.Dividends = append(pinfo.Dividends,
				&schema.Dividend{
					Date:  date,
					Value: op.Payment,
				})
		} else {
			log.Printf("Unprocessed transaction %v", op)
		}
	}

	for _, pinfo := range p.positions {
		sort.Slice(pinfo.Deals, func(i, j int) bool {
			return pinfo.Deals[i].Date.Before(pinfo.Deals[j].Date)
		})
		p.makePortions(pinfo)
	}

	return nil
}

func (p *Portfolio) makePortions(pinfo *schema.PositionInfo) {
	var po *schema.Portion
	var balance int
	var spent float64

	now := time.Now()

	for _, deal := range pinfo.Deals {
		balance += deal.Quantity
		spent += deal.Value()

		if deal.Quantity > 0 { // buy
			if po == nil {
				// first deal
				po = &schema.Portion{
					Balance: schema.NewCValue(0, deal.Price.Currency),
					AvgDate: deal.Date,
				}
			} else {
				// TODO think again is this correct?
				mult := deal.Value() / spent

				biasDays := int(math.Round(deal.Date.Sub(po.AvgDate).Hours() *
					mult / 24))
				po.AvgDate.AddDate(0, 0, biasDays)

				po.AvgPrice.Value = deal.Price.Value*mult +
					po.AvgPrice.Value*(1-mult)
			}

			po.Buys = append(po.Buys, deal)

		} else { // sell
			if balance > 0 {
				log.Printf("Partial sells are not handled nicely yet")
				po = nil
				break
			}
			if balance < 0 {
				log.Fatal("wat")
			}

			// complete sell
			po.Close = deal
			po.IsClosed = true
			pinfo.Portions = append(pinfo.Portions, po)
			// begin to fill new portion
			po = nil
		}
	}

	if po != nil {
		if pinfo.IsClosed {
			log.Fatal("wat")
		}

		po.Close = &schema.Deal{
			Date:     now,
			Price:    pinfo.CurrentPrice,
			Quantity: -balance,
		}
		pinfo.Portions = append(pinfo.Portions, po)
	}

	// can now calculate balance and yields
	for _, po = range pinfo.Portions {
		var expense float64

		profit := po.Close.Price.Mult(float64(-po.Close.Quantity))

		for _, div := range pinfo.Dividends {
			if div.Date.Before(po.Buys[0].Date) {
				continue
			}
			if div.Date.After(po.Close.Date) {
				// TODO not quite right. Dividends come with delay
				continue
			}
			profit.Value += div.Value
		}

		for _, deal := range po.Buys {
			expense += deal.Value()
			expense += -deal.Commission
		}
		expense += -po.Close.Commission

		po.Yield = profit.Div(expense / 100)
		po.Yield.Value -= 100

		po.YieldAnnual = po.Yield.Value * 365 / (po.Close.Date.Sub(po.AvgDate).Hours() / 24)

		po.Balance = profit
		po.Balance.Value -= expense
	}
}
