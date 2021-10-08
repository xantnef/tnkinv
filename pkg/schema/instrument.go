package schema

import (
	"time"

	log "github.com/sirupsen/logrus"

	"strings"

	"../aux"
)

type InsType string

const (
	InsTypeEtf      InsType = "Etf"
	InsTypeBond             = "Bond"
	InsTypeStock            = "Stock"
	InsTypeCurrency         = "Currency"
)

type Section string

const (
	BondRu  Section = "Bond.RU"
	BondUs          = "Bond.US"
	StockRu         = "Stock.RU"
	StockEm         = "Stock.EM"
	StockUs         = "Stock.US"
	StockDm         = "Stock.DM"
	CashRu          = "Cash.RU"
	CashUs          = "Cash.US"
)

// TODO why json tags?
type Instrument struct {
	Figi      string `json:"figi"`
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	Lot       int
	FaceValue int `json:"faceValue"`

	Type    InsType
	Section Section
}

func NewInstrument(figi, ticker, name, typ, currency string, faceValue, lot int) Instrument {
	if !Currencies.Has(currency) {
		log.Fatalf("unknown currency %s (%s)", currency, ticker)
	}

	ins := Instrument{
		Figi:      figi,
		Ticker:    ticker,
		Name:      name,
		Currency:  currency,
		FaceValue: faceValue,
		Lot:       lot,
	}
	ins.Type = getInstrumentType(typ, ticker)
	ins.Section = getSection(ins)
	return ins
}

func getInstrumentType(typ string, ticker string) InsType {
	if !map[InsType]bool{
		InsTypeEtf:      true,
		InsTypeStock:    true,
		InsTypeBond:     true,
		InsTypeCurrency: true,
	}[InsType(typ)] {
		log.Warnf("Unhandled type %s: %s", ticker, typ)
	}
	return InsType(typ)
}

func splitCoef(ticker string, date time.Time) int {
	if aux.IsIn(ticker, "VTBB", "VTBE") && date.Before(time.Date(2021, 4, 12, 0, 0, 0, 0, time.UTC)) {
		return 10
	}
	if aux.IsIn(ticker, "FXDE") && date.Before(time.Date(2021, 9, 7, 0, 0, 0, 0, time.UTC)) {
		return 100
	}
	return 1
}

func (ins Instrument) Benchmark() string {
	if bench, ok := map[Section]string{
		BondRu:  "VTBB",
		BondUs:  "FXRU",
		StockRu: "FXRL",
		StockEm: "VTBE",
		// StockDm: "FXDM", // would be nice, but it appeared too recently; cant compare early dates
		StockDm: "FXUS",
		StockUs: "FXUS", // see below
	}[ins.Section]; ok {
		if bench == "FXUS" && aux.IsIn(ins.Ticker,
			"AAPL", // 18.1%
			"MSFT", // 16.3%
			"GOOG", //  9.6%
			"FB",   //  6.2%
			"V",    //  3.6%
			"MA",   //  2.9%
			"INTC", //  2.8%
			"NVDA", //  2.6%
			"NFLX", //  2.3%
		) {
			return "FXIT"
		}
		if bench != ins.Ticker {
			return bench
		}
	}

	return ""
}

func (s Section) Currency() string {
	if strings.HasSuffix(string(s), ".RU") {
		return "RUB"
	}
	return "USD"
}

func GetEtfSection(ticker string) (Section, bool) {
	s, ok := map[string]Section{
		"VTBB": BondRu,
		"FXRB": BondRu,
		"TBRU": BondRu,

		// T* funds are (25x4 gold, stocks, long and short bonds)
		// TODO proper accounting
		// consider them bonds for now
		"TRUR": BondRu,
		"TUSD": BondUs,

		"FXRU": BondUs,
		"VTBU": BondUs,

		"SBMX": StockRu,
		"FXRL": StockRu,
		"TMOS": StockRu,

		"AKNX": StockUs,
		"FXIT": StockUs,
		"FXIM": StockUs,
		"FXUS": StockUs,

		"FXDM": StockDm,
		"FXDE": StockDm,

		"TECH": StockUs,
		"TSPX": StockUs,
		"TIPO": StockUs,
		"TBIO": StockUs,

		"VTBE": StockEm,
		"FXCN": StockEm,

		"FXMM": CashRu,
		"FXTB": CashUs,
	}[ticker]
	return s, ok
}

func getSection(ins Instrument) Section {
	if s, ok := map[string]Section{
		InsTypeBond + "RUB": BondRu,
		InsTypeBond + "USD": BondUs,

		InsTypeStock + "RUB": StockRu,
		InsTypeStock + "USD": StockUs,

		InsTypeCurrency + "RUB": CashRu,
		InsTypeCurrency + "USD": CashUs,
	}[string(ins.Type)+ins.Currency]; ok {
		return s
	}

	if ins.Type == InsTypeEtf {
		s, ok := GetEtfSection(ins.Ticker)
		if !ok {
			log.Warnf("Uncatched ETF %s", ins.Ticker)
		}
		return s
	}

	return ""
}
