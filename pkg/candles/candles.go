package candles

import (
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"../client"
	"../schema"
)

type CandleCache struct {
	client *client.MyClient
	start  time.Time
	period string

	cache map[string]*schema.CandlesResponse
}

func NewCandleCache(c *client.MyClient, start time.Time, period string) *CandleCache {
	return &CandleCache{
		client: c,
		start:  start,
		period: period,

		cache: make(map[string]*schema.CandlesResponse),
	}
}

func (cc *CandleCache) List(figi string) *schema.CandlesResponse {
	pcandles := cc.cache[figi]
	if pcandles != nil {
		return pcandles
	}

	resp := cc.client.RequestCandles(figi, cc.start, time.Now(), cc.period)
	pcandles = &resp

	// TODO
	// Candles for amortized bonds are quire fucked up.
	// They show the old values as if it was now.
	// e.g.
	//   1. price = 1000
	//   2. grew to 1100
	//   3. amortized to 880
	// candle for the time point (1) is going to show 800.

	for i := range pcandles.Payload.Candles {
		var err error
		c := &pcandles.Payload.Candles[i]

		c.TimeParsed, err = time.Parse(time.RFC3339, c.Time)
		if err != nil {
			log.Fatal("failed to parse time %v", err)
		}
	}

	sort.Slice(pcandles.Payload.Candles, func(i, j int) bool {
		return pcandles.Payload.Candles[i].TimeParsed.Before(pcandles.Payload.Candles[j].TimeParsed)
	})

	cc.cache[figi] = pcandles
	return pcandles
}

func (cc *CandleCache) get(figi string, t time.Time) (*schema.Candle, bool) {
	pcandles := cc.List(figi).Payload.Candles

	idx := sort.Search(len(pcandles), func(i int) bool {
		return pcandles[i].TimeParsed.Equal(t) || pcandles[i].TimeParsed.After(t)
	})

	if idx == len(pcandles) {
		return &pcandles[idx-1], true
	}

	return &pcandles[idx], false
}

func (cc *CandleCache) Get(figi string, t time.Time) float64 {
	c, last := cc.get(figi, t)
	if last {
		return c.C
	} else {
		return c.O
	}
}

func (cc *CandleCache) Pricef(t time.Time) func(figi string) float64 {
	return func(figi string) float64 {
		return cc.Get(figi, t)
	}
}
