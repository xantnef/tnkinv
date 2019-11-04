package myclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"time"

	swagger "../go-client"
	"../schema"
)

type Config struct {
	Token     string
	SandToken string
}

type MyClient interface {
	Run() error
	Stop()
}

type myClient struct {
	swc       *swagger.APIClient
	cfg       Config
	portfolio portfolio
}

type portfolio struct {
	tickers   map[string]string
	positions map[string]*schema.PositionInfo

	figisSorted []string

	totals struct {
		commission schema.CValue
		payins     map[string]*schema.CValue
		assets     map[string]*schema.CValue
	}
}

func NewClient(cfg *Config) MyClient {
	c := &myClient{
		portfolio: newPortfolio(),
	}

	if cfg != nil {
		c.cfg = *cfg
	}

	return c
}

func newPortfolio() portfolio {
	p := portfolio{
		tickers:   make(map[string]string),
		positions: make(map[string]*schema.PositionInfo),
	}

	p.totals.payins = make(map[string]*schema.CValue)
	p.totals.assets = make(map[string]*schema.CValue)
	p.makeCurrencies()

	return p
}

func (p *portfolio) makeCurrencies() {
	for _, m := range []map[string]*schema.CValue{p.totals.payins, p.totals.assets} {
		for cur := range schema.Currencies {
			cv := schema.NewCValue(0, cur)
			m[cur] = &cv
		}
	}
	p.totals.commission.Currency = "RUB"
}

func (c *myClient) getToken(fname string) string {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func (c *myClient) runSandbox() error {
	conf := swagger.NewConfiguration()
	conf.BasePath = "https://api-invest.tinkoff.ru/openapi/sandbox/"
	conf.AddDefaultHeader("Authorization", "Bearer "+c.getToken(c.cfg.SandToken))

	swc := swagger.NewAPIClient(conf)

	sand := swc.SandboxApi

	_, err := sand.SandboxRegisterPost(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("sandbox register complete")
	return nil
}

func (c *myClient) getAPI() *swagger.APIClient {
	if c.swc == nil {
		conf := swagger.NewConfiguration()
		conf.BasePath = "https://api-invest.tinkoff.ru/openapi/"
		conf.AddDefaultHeader("Authorization", "Bearer "+c.getToken(c.cfg.Token))

		c.swc = swagger.NewAPIClient(conf)
	}
	return c.swc
}

func (c *myClient) requestCurrentPrice(figi string) float64 {
	mktApi := c.getAPI().MarketApi
	mktResp := schema.OrderbookResponse{}

	body, err := mktApi.MarketOrderbookGet(nil, figi, 1)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &mktResp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))

	return mktResp.Payload.LastPrice
}

func (c *myClient) requestTicker(figi string) string {
	mktApi := c.getAPI().MarketApi
	resp := schema.SearchByFigiResponse{}

	body, err := mktApi.MarketSearchByFigiGet(nil, figi)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal(err)
	}

	return resp.Payload.Ticker
}

func (c *myClient) requestPortfolio() schema.PortfolioResponse {
	pfApi := c.getAPI().PortfolioApi
	pfResp := schema.PortfolioResponse{}

	body, err := pfApi.PortfolioGet(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &pfResp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))
	//log.Print(pfResp)

	return pfResp
}

func (c *myClient) requestOperations() schema.OperationsResponse {
	//2019-08-19T18:38:33.131642+03:00
	timeStartStr := time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	timeNow := time.Now()

	opsApi := c.getAPI().OperationsApi
	opsResp := schema.OperationsResponse{}

	body, err := opsApi.OperationsGet(nil, timeStartStr, timeNow.Format(time.RFC3339), nil)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &opsResp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))
	//log.Print(opsResp)

	return opsResp
}

func (p *portfolio) processPortfolio(c *myClient) error {
	pfResp := c.requestPortfolio()

	for _, pos := range pfResp.Payload.Positions {
		currency := pos.ExpectedYield.Currency
		p.tickers[pos.Figi] = pos.Ticker

		pinfo := &schema.PositionInfo{
			Figi:         pos.Figi,
			Ticker:       pos.Ticker,
			CurrentPrice: schema.NewCValue(c.requestCurrentPrice(pos.Figi), currency),
			Quantity:     pos.Balance,

			AccumulatedIncome: schema.NewCValue(0, currency),
		}

		p.positions[pos.Figi] = pinfo

		if pos.Figi == schema.FigiUSD {
			p.totals.assets["USD"].Value += pos.Balance
		} else {
			p.totals.assets[currency].Value += pinfo.CurrentPrice.Value * pinfo.Quantity
		}
	}

	opsResp := c.requestOperations()

	//log.Print("== Transaction log ==")

	for _, op := range opsResp.Payload.Operations {
		date, err := time.Parse(time.RFC3339, op.Date)
		if err != nil {
			// crutch one crippled transaction
			if op.Date == "2019-08-23T00:00+03:00" {
				op.Date = "2019-08-23T00:00:00+03:00"
			} else {
				log.Printf("Failed to parse time: %v", err)
			}
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
				p.totals.payins[op.Currency].Value += op.Payment
			} else if op.OperationType == "ServiceCommission" {
				p.totals.commission.Value += op.Payment
			}
			continue
		}

		pinfo := p.positions[op.Figi]
		if pinfo == nil {
			pinfo = &schema.PositionInfo{
				Figi:     op.Figi,
				Ticker:   c.requestTicker(op.Figi),
				IsClosed: true,
			}
			p.positions[op.Figi] = pinfo
		}

		if op.OperationType == "Buy" || op.OperationType == "BuyCard" || op.OperationType == "Sell" {
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
		p.makePortions(pinfo)
	}

	return nil
}

func (p *portfolio) makePortions(pinfo *schema.PositionInfo) {
	var po *schema.Portion
	var balance int

	now := time.Now()

	sort.Slice(pinfo.Deals, func(i, j int) bool {
		return pinfo.Deals[i].Date.Before(pinfo.Deals[j].Date)
	})

	for _, deal := range pinfo.Deals {

		balance += deal.Quantity

		if deal.Quantity > 0 { // buy
			if po == nil {
				// first deal
				po = &schema.Portion{
					Balance: schema.NewCValue(0, deal.Price.Currency),
					AvgDate: deal.Date,
				}
				pinfo.Portions = append(pinfo.Portions, po)
			} else {
				// TODO this is wrong
				mult := float64(deal.Quantity) / float64(deal.Quantity+balance)

				biasDays := int(math.Round(deal.Date.Sub(po.AvgDate).Hours() * mult / 24))
				po.AvgDate.AddDate(0, 0, biasDays)

				po.AvgPrice.Value = deal.Price.Value*mult + po.AvgPrice.Value*(1-mult)
			}

			po.Buys = append(po.Buys, deal)

		} else { // sell
			if balance > 0 {
				log.Printf("Partial sells are not handled nicely yet")
				break
			}
			if balance < 0 {
				log.Fatal("wat")
			}

			// complete sell
			po.Close = deal
			po.IsClosed = true
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
	}

	// can now calculate balance and yields
	for _, po = range pinfo.Portions {
		var expense float64 = 0

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
			expense += deal.Price.Value * float64(deal.Quantity)
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

func (c *myClient) Run() error {
	if c.cfg.SandToken != "" {
		c.runSandbox()
	}

	if c.cfg.Token == "" {
		log.Printf("no token provided")
		return nil
	}

	c.portfolio.processPortfolio(c)
	c.portfolio.print()

	return nil
}

func (c *myClient) Stop() {

}
