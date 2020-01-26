package client

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	swagger "../go-client"
	"../schema"
)

type MyClient struct {
	swc    *swagger.APIClient
	tokenf string
}

func NewClient(tokenf string) *MyClient {
	return &MyClient{
		tokenf: tokenf,
	}
}

func (c *MyClient) getToken(fname string) string {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func (c *MyClient) TrySandbox() error {
	conf := swagger.NewConfiguration()
	conf.BasePath = "https://api-invest.tinkoff.ru/openapi/sandbox/"
	conf.AddDefaultHeader("Authorization", "Bearer "+c.getToken(c.tokenf))

	swc := swagger.NewAPIClient(conf)

	sand := swc.SandboxApi

	_, err := sand.SandboxRegisterPost(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("sandbox register complete")
	return nil
}

func (c *MyClient) getAPI() *swagger.APIClient {
	if c.swc == nil {
		conf := swagger.NewConfiguration()
		conf.BasePath = "https://api-invest.tinkoff.ru/openapi/"
		conf.AddDefaultHeader("Authorization", "Bearer "+c.getToken(c.tokenf))

		c.swc = swagger.NewAPIClient(conf)
	}
	return c.swc
}

func (c *MyClient) RequestCurrentPrice(figi string) float64 {
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

func (c *MyClient) RequestTicker(figi string) string {
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

func (c *MyClient) RequestFigi(ticker string) string {
	mktApi := c.getAPI().MarketApi
	resp := schema.SearchByTickerResponse{}

	body, err := mktApi.MarketSearchByTickerGet(nil, ticker)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))
	//log.Print(resp)

	return resp.Payload.Instruments[0].Figi
}

func (c *MyClient) RequestPortfolio() schema.PortfolioResponse {
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

func (c *MyClient) RequestOperations(start time.Time) schema.OperationsResponse {
	timeStartStr := start.Format(time.RFC3339)
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

func (c *MyClient) RequestCandles(figi string, t1, t2 time.Time, interval string) schema.CandlesResponse {
	t1Str := t1.Format(time.RFC3339)
	t2Str := t2.Format(time.RFC3339)

	mktApi := c.getAPI().MarketApi
	mktResp := schema.CandlesResponse{}

	body, err := mktApi.MarketCandlesGet(nil, figi, t1Str, t2Str, interval)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &mktResp)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(string(body))
	//log.Print(mktResp)

	return mktResp
}

func (c *MyClient) Stop() {

}
