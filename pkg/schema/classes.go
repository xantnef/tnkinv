package schema

import (
	log "github.com/sirupsen/logrus"

	"strings"

	"../aux"
)

func GetInstrumentType(typ string, ticker string) InsType {
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

func GetBenchmark(ins Instrument) string {
	if bench, ok := map[string]string{
		InsTypeBond + "RUB":  "VTBB",
		InsTypeBond + "USD":  "FXRU",
		InsTypeStock + "RUB": "FXRL",
		InsTypeStock + "USD": "FXUS", // see below
	}[string(ins.Type)+ins.Currency]; ok {

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
		return bench
	}

	if ins.Type == InsTypeEtf {
		// ETF is benchmark itself
	}

	return ""
}

func GetSectionCurrency(s Section) string {
	for cur := range Currencies {
		if strings.Contains(string(s), cur) {
			return cur
		}
	}
	return ""
}

func GetEtfSection(ticker string) (Section, bool) {
	s, ok := map[string]Section{
		"VTBB": BondRub,
		"FXRB": BondRub,

		// T* funds are (25x4 gold, stocks, long and short bonds)
		// TODO proper accounting
		// consider them bonds for now
		"TRUR": BondRub,
		"TUSD": BondUsd,

		"FXRU": BondUsd,

		"SBMX": StockRub,
		"FXRL": StockRub,
		"TMOS": StockRub,

		"AKNX": StockUsd,
		"FXIT": StockUsd,
		"FXUS": StockUsd,
		// it's actually StockEur, but leave it for now
		"FXDE": StockUsd,
		"TECH": StockUsd,

		"FXMM": CashRub,
		"FXTB": CashUsd,
	}[ticker]
	return s, ok
}

func GetSection(ins Instrument) Section {
	if s, ok := map[string]Section{
		InsTypeBond + "RUB": BondRub,
		InsTypeBond + "USD": BondUsd,

		InsTypeStock + "RUB": StockRub,
		InsTypeStock + "USD": StockUsd,

		InsTypeCurrency + "RUB": CashRub,
		InsTypeCurrency + "USD": CashUsd,
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
