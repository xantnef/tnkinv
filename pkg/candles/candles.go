package candles

import (
	"errors"
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

	periodFetched map[string]bool            // key=figi
	cache         map[string][]schema.Candle // key=figi
}

func NewCandleCache(c *client.MyClient, start time.Time, period string) *CandleCache {
	return &CandleCache{
		client: c,
		start:  start,
		period: period,

		periodFetched: make(map[string]bool),
		cache:         make(map[string][]schema.Candle),
	}
}

func (cc *CandleCache) fetchCandles(figi string, start, end time.Time, period string) []schema.Candle {
	resp := cc.client.RequestCandles(figi, start, end, period)
	pcandles := resp.Payload.Candles

	// TODO
	// Candles for amortized bonds are quire fucked up.
	// They show the old values as if it was now.
	// e.g.
	//   1. price = 1000
	//   2. grew to 1100
	//   3. amortized to 880
	// candle for the time point (1) is going to show 800.

	if len(pcandles) == 0 {
		log.Infof("No candles for period %s - %s", start, end)
		return cc.fetchCandles(figi, start.Add(-24*time.Hour), end, period)
	}

	for i := range pcandles {
		var err error
		c := &pcandles[i]

		c.TimeParsed, err = time.Parse(time.RFC3339, c.Time)
		if err != nil {
			log.Fatal("failed to parse time %v", err)
		}
	}

	return pcandles
}

func sortCandles(pcandles []schema.Candle) []schema.Candle {
	sort.Slice(pcandles, func(i, j int) bool {
		return pcandles[i].TimeParsed.Before(pcandles[j].TimeParsed)
	})
	return pcandles
}

func (cc *CandleCache) fetchPeriod(figi string) {
	if cc.periodFetched[figi] {
		return
	}
	cc.periodFetched[figi] = true

	now := time.Now()

	pcandles := cc.fetchCandles(figi, cc.start, now, cc.period)
	cc.cache[figi] = sortCandles(append(cc.cache[figi], pcandles...))
}

func (cc *CandleCache) findInCache(figi string, t time.Time, exact bool) (float64, error) {
	logAndRet := func(c schema.Candle, isOpening bool) (float64, error) {
		if isOpening {
			log.Debugf("price(%s on %s): returning opening on %s = %f",
				figi, t, c.Time, c.O)
			return c.O, nil
		} else {
			log.Debugf("price(%s on %s): returning closing on %s = %f",
				figi, t, c.Time, c.C)
			return c.C, nil
		}
	}

	pcandles, exist := cc.cache[figi]
	if !exist {
		return 0, errors.New("no cache")
	}

	idx := sort.Search(len(pcandles), func(i int) bool {
		return pcandles[i].TimeParsed.Year() == t.Year() &&
			pcandles[i].TimeParsed.YearDay() == t.YearDay()
	})

	if idx < len(pcandles) {
		return logAndRet(pcandles[idx], false)
	}

	if exact {
		return 0, errors.New("exact date not found")
	}

	idx = sort.Search(len(pcandles), func(i int) bool {
		return pcandles[i].TimeParsed.After(t)
	})

	if idx == 0 {
		// all candles are after
		return logAndRet(pcandles[0], true)
	}

	if idx == len(pcandles) {
		// all candles are before
		return logAndRet(pcandles[len(pcandles)-1], false)
	}

	if t.Sub(pcandles[idx-1].TimeParsed) < pcandles[idx].TimeParsed.Sub(t) {
		return logAndRet(pcandles[idx-1], false)
	} else {
		return logAndRet(pcandles[idx], true)
	}
}

func (cc *CandleCache) find(figi string, t time.Time) float64 {
	price, err := cc.findInCache(figi, t, false)
	if err != nil {
		log.Fatalf("Couldnt get candles? %s %s: %s", figi, t, err)
	}
	return price
}

func (cc *CandleCache) findExactDate(figi string, t time.Time) (float64, error) {
	return cc.findInCache(figi, t, true)
}

func (cc *CandleCache) Get(figi string, t time.Time) float64 {
	cc.fetchPeriod(figi)
	return cc.find(figi, t)
}

func (cc *CandleCache) GetOnDay(figi string, t time.Time) float64 {
	price, err := cc.findExactDate(figi, t)
	if err == nil {
		return price
	}

	// normalize the time a bit
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// we wont catch all the holidays here, but it covers 95%
	wday := start.Weekday()
	if wday == time.Saturday {
		start = start.Add(-1 * 24 * time.Hour)
	} else if wday == time.Sunday {
		start = start.Add(-2 * 24 * time.Hour)
	}

	pcandles := cc.fetchCandles(figi, start, start.Add(24*time.Hour), "day")

	// might be a duplicate. Nothing critical
	cc.cache[figi] = sortCandles(append(cc.cache[figi], pcandles...))

	return cc.find(figi, t)
}

func (cc *CandleCache) ListTimes() (times []time.Time) {
	cc.fetchPeriod(schema.FigiUSD)
	for _, c := range cc.cache[schema.FigiUSD] {
		times = append(times, c.TimeParsed)
	}
	return
}

func (cc *CandleCache) Pricef1(t time.Time) schema.Pricef1 {
	return func(figi string) float64 {
		return cc.Get(figi, t)
	}
}
