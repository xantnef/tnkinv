package myclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	swc *swagger.APIClient

	cfg Config

	tickers   map[string]string
	positions map[string]*schema.PositionInfo

	totals struct {
		commission schema.CValue
		payins     map[string]*schema.CValue
		assets     map[string]*schema.CValue
	}
}

func NewClient(cfg *Config) (MyClient, error) {
	c := &myClient{
		tickers:   make(map[string]string),
		positions: make(map[string]*schema.PositionInfo),
	}

	c.totals.payins = make(map[string]*schema.CValue)
	c.totals.assets = make(map[string]*schema.CValue)
	c.makeCurrencies()

	if cfg != nil {
		c.cfg = *cfg
	}

	return c, nil
}

func (c *myClient) makeCurrencies() {
	for _, cur := range []string{"RUB", "USD"} {
		for _, m := range []map[string]*schema.CValue{c.totals.payins, c.totals.assets} {
			cv := schema.NewCValue(0, cur)
			m[cur] = &cv
		}
	}
	c.totals.commission.Currency = "RUB"
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

func (c *myClient) processPortfolio() error {
	//2019-08-19T18:38:33.131642+03:00
	timeStartStr := time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	timeNow := time.Now()

	pfApi := c.getAPI().PortfolioApi
	pfResp := schema.PortfolioResponse{}

	opsApi := c.getAPI().OperationsApi
	opsResp := schema.OperationsResponse{}

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

	for _, pos := range pfResp.Payload.Positions {
		currency := pos.ExpectedYield.Currency
		c.tickers[pos.Figi] = pos.Ticker

		pinfo := &schema.PositionInfo{
			Figi:         pos.Figi,
			Ticker:       pos.Ticker,
			CurrentPrice: schema.NewCValue(c.requestCurrentPrice(pos.Figi), currency),
			Quantity:     pos.Balance,

			AccumulatedIncome: schema.NewCValue(0, currency),
		}

		c.positions[pos.Figi] = pinfo
		c.totals.assets[currency].Value += pinfo.CurrentPrice.Value * pinfo.Quantity
	}

	body, err = opsApi.OperationsGet(nil, timeStartStr, timeNow.Format(time.RFC3339), nil)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &opsResp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))
	//log.Print(opsResp)

	//log.Print("== Transaction log ==")

	for _, op := range opsResp.Payload.Operations {
		date, err := time.Parse(time.RFC3339, op.Date)
		if err != nil {
			log.Printf("Failed to parse time: %v", err)
		}

		/* log.Printf("at %s %s some %s",
		date.String(), op.OperationType+"-ed",
		c.tickers[op.Figi])*/

		if op.Status != "Done" {
			// cancelled declined etc
			continue
		}

		if op.Figi == "" {
			if op.OperationType == "PayIn" {
				c.totals.payins[op.Currency].Value += op.Payment
			} else if op.OperationType == "ServiceCommission" {
				c.totals.commission.Value += op.Payment
			}
			continue
		}

		pinfo := c.positions[op.Figi]
		if pinfo == nil {
			log.Printf("Unhandled sold asset %s", op.Figi)
			// TODO
			continue
		}

		if op.OperationType == "Buy" || op.OperationType == "BuyCard" {
			//log.Printf("%f:%f", pinfo.CurrentPrice.Value, op.Price)
			deal := schema.Deal{
				Opened:   date,
				Price:    schema.NewCValue(op.Price, op.Currency),
				Quantity: op.Quantity,
				Yield:    schema.NewCValue(pinfo.CurrentPrice.Value/op.Price*100-100, op.Currency),
			}
			deal.YieldAnnual = deal.Yield.Value * 365 / (timeNow.Sub(date).Hours() / 24)
			pinfo.Deals = append(pinfo.Deals, deal)

		} else if op.OperationType == "Sell" {
			log.Printf("Unhandled sell operation for %s", pinfo.Ticker)
		} else if op.OperationType == "BrokerComission" {
			pinfo.AccumulatedIncome.Value += op.Payment
		} else if op.OperationType == "Dividend" || op.OperationType == "TaxDividend" {
			pinfo.AccumulatedIncome.Value += op.Payment
		}
	}

	return nil
}

func (c *myClient) Run() error {
	if c.cfg.SandToken != "" {
		c.runSandbox()
	}

	if c.cfg.Token == "" {
		log.Printf("no token provided")
		return nil
	}

	c.processPortfolio()
	c.printPortfolio()

	return nil
}

func (c *myClient) Stop() {

}
