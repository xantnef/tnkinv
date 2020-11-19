package portfolio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
	var totalAmount float64

	fs := readFictives(fname)

	for _, op := range fs {
		totalAmount += op.Amount
	}

	for _, op := range fs {
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
		n := uint(math.Round(op.Amount/(price*float64(ins.Lot)))) * uint(ins.Lot)
		pment := price * float64(n)

		s := fmt.Sprintf("%-5s spent %8.2f/%8.2f - %5.1f%% (%5.1f%% -> %5.1f%%)",
			ins.Ticker, pment, op.Amount, 100*pment/op.Amount,
			100*op.Amount/totalAmount, 100*pment/totalAmount)
		if n == 0 {
			s += " " + fmt.Sprintf("(min %8.2f)", float64(ins.Lot)*price)
		}
		log.Info(s)

		ops = append(ops,
			schema.Operation{
				Date:           date.Format(time.RFC3339),
				Figi:           ins.Figi,
				InstrumentType: string(ins.Type),

				Price:     price,
				Currency:  op.Currency,
				Quantity_: n,
				Payment:   -pment,

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
				Payment:  pment,
				Currency: op.Currency,

				OperationType: "PayIn",
				Status:        "Done",
			},
		)
	}

	return
}
