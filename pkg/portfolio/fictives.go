package portfolio

import (
	"encoding/json"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"

	"../candles"
	"../client"
	"../schema"
)

type FictiveDeal struct {
	Ticker   string
	Date     string
	Amount   float64
	Currency string
}

func readFictives(fname string) (ops []FictiveDeal) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &ops)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func fetchFictives(c *client.MyClient, cc *candles.CandleCache, fname string) (ops []schema.Operation) {

	for _, op := range readFictives(fname) {
		date, err := time.Parse("2006/01/02", op.Date)
		if err != nil {
			log.Errorf("bad date %s: %s", op.Ticker, op.Date)
			return
		}

		if !schema.Currencies.Has(op.Currency) {
			log.Errorf("bad currency %s: %s", op.Ticker, op.Currency)
			return
		}

		ins, err := c.TryRequestByTicker(op.Ticker)
		if err != nil {
			log.Errorf("bad ticker %s: %s", op.Ticker, err)
			return
		}

		price := cc.Get(ins.Figi, date)

		n := uint(op.Amount / price)
		ops = append(ops,
			schema.Operation{
				Date:           date.Format(time.RFC3339),
				Figi:           ins.Figi,
				InstrumentType: string(ins.Type),

				Price:     price,
				Currency:  op.Currency,
				Quantity_: n,
				Payment:   -price * float64(n),

				OperationType: "Buy",
				Status:        "Done",

				Trades: []schema.Trade{
					schema.Trade{
						Date:     date.Format(time.RFC3339),
						Price:    price,
						Quantity: n,
					},
				},
			},
			schema.Operation{
				Date:     date.Format(time.RFC3339),
				Payment:  price * float64(n),
				Currency: op.Currency,

				OperationType: "PayIn",
				Status:        "Done",
			},
		)
	}
	return
}
