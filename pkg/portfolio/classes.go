package portfolio

import (
	log "github.com/sirupsen/logrus"

	"../aux"
	"../schema"
)

func getInstrumentType(typ string, ticker string) schema.InsType {
	m := map[schema.InsType]bool{
		schema.InsTypeEtf:      true,
		schema.InsTypeStock:    true,
		schema.InsTypeBond:     true,
		schema.InsTypeCurrency: true,
	}
	if !m[schema.InsType(typ)] {
		log.Warnf("Unhandled type %s: %s", ticker, typ)
	}
	return schema.InsType(typ)
}

func getBenchmark(ticker string, typ schema.InsType, currency string) string {
	m := map[string]string{
		schema.InsTypeBond + "RUB":  "VTBB",
		schema.InsTypeBond + "USD":  "FXRU",
		schema.InsTypeStock + "RUB": "FXRL",
		schema.InsTypeStock + "USD": "FXUS", // see below
	}

	if bench, ok := m[string(typ)+currency]; ok {
		if bench == "FXUS" {
			if aux.IsIn(ticker,
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
		}
		return bench
	}

	if typ == schema.InsTypeEtf {
		// ETF is benchmark itself
	}

	return ""
}

func getEtfSection(ticker string) (schema.Section, bool) {
	m := map[string]schema.Section{
		"VTBB": schema.BondRub,
		"FXRB": schema.BondRub,

		// T* funds are (25x4 gold, stocks, long and short bonds)
		// TODO proper accounting
		// consider them bonds for now
		"TRUR": schema.BondRub,
		"TUSD": schema.BondUsd,

		"FXRU": schema.BondUsd,

		"SBMX": schema.StockRub,
		"FXRL": schema.StockRub,

		"AKNX": schema.StockUsd,
		"FXIT": schema.StockUsd,
		"FXUS": schema.StockUsd,
		// it's actually StockEur, but leave it for now
		"FXDE": schema.StockUsd,

		"FXMM": schema.CashRub,
		"FXTB": schema.CashUsd,
	}
	s, ok := m[ticker]
	return s, ok
}

func getSection(ticker string, typ schema.InsType, currency string) schema.Section {
	m := map[string]schema.Section{
		schema.InsTypeBond + "RUB": schema.BondRub,
		schema.InsTypeBond + "USD": schema.BondUsd,

		schema.InsTypeStock + "RUB": schema.StockRub,
		schema.InsTypeStock + "USD": schema.StockUsd,

		schema.InsTypeCurrency + "RUB": schema.CashRub,
		schema.InsTypeCurrency + "USD": schema.CashUsd,
	}
	if s, ok := m[string(typ)+currency]; ok {
		return s
	}

	if typ == schema.InsTypeEtf {
		s, ok := getEtfSection(ticker)
		if !ok {
			log.Warnf("Uncatched ETF %s", ticker)
		}
		return s
	}

	return ""
}
