package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"../pkg/aux"
	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token, sideOps, fictOps, period, format, acc string

	tickers []string

	start, end, at time.Time

	startSet bool
}

func parseDate(s string, def time.Time) (time.Time, bool) {
	if s == "" {
		return def, false
	}

	t, err := time.Parse("2006/01/02", s)
	if err != nil {
		usage()
		log.Fatalf("unrecognized date %s", s)
	}

	return t, true
}

func parseCmdline() (string, config) {
	if len(os.Args) < 2 {
		usage()
		log.Fatal("no cmd provided")
	}

	cfg := config{}

	// --------------
	// Verify command

	cmd := os.Args[1]
	cmds := aux.NewList(
		"sandbox",
		"show",
		"story",
		"deals",
		"price",
	)

	if !cmds.Has(cmd) {
		usage()
		log.Fatalf("unknown command %s", cmd)
	}

	// ------------
	// List options

	fs := flag.NewFlagSet("", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	sideOps := fs.String("operations", "", "json file with operations")
	fictOps := fs.String("fictives", "", "json file with fictive operations")
	acc := fs.String("account", "broker", "account")
	loglevel := fs.String("loglevel", "none", "log level")

	period := fs.String("period", "", "story period")
	start := fs.String("start", "", "starting point in time (format: 1922/12/28; default: year ago)")
	end := fs.String("end", "", "end point in time (format: 1922/12/28; default: now)")
	atTime := fs.String("at", "", "point in time (default: now). Not supported yet")
	format := fs.String("format", "human", "output format")
	tickers := fs.String("tickers", "", "list of tickers")

	fs.Parse(os.Args[2:])

	cfg.token = *token
	cfg.sideOps = *sideOps
	cfg.fictOps = *fictOps
	if *tickers != "" {
		cfg.tickers = strings.Split(*tickers, ",")
	}

	// ----------------
	// Verify log level

	loglevels := map[string]log.Level{
		"none":  log.InfoLevel,
		"debug": log.DebugLevel,
		"all":   log.TraceLevel,
	}
	if _, ok := loglevels[*loglevel]; !ok {
		log.Fatalf("bad log level %s", *loglevel)
	}

	log.SetLevel(loglevels[*loglevel])

	// -------------
	// Verify format

	formats := aux.NewList(
		"human",
		"table",
	)
	if !formats.Has(*format) {
		log.Fatalf("bad format %s", *format)
	}
	cfg.format = *format

	// --------------
	// Verify account

	accs := aux.NewList(
		"broker",
		"iis",
		"all",
	)
	if !accs.Has(*acc) {
		log.Fatalf("bad account type %s", *acc)
	}
	cfg.acc = *acc

	// --------------
	// Verify period

	pers := aux.NewList(
		"",
		"day",
		"week",
		"month",
		"all",
	)
	if !pers.Has(*period) {
		log.Fatalf("bad period %s", *period)
	}
	cfg.period = *period

	// ----------------------
	// Parse and verify times

	cfg.start, cfg.startSet = parseDate(*start, time.Now().AddDate(-1, 0, 0))
	cfg.end, _ = parseDate(*end, time.Now())
	cfg.at, _ = parseDate(*atTime, time.Now())

	return cmd, cfg
}

func usage() {
	fmt.Printf("usage:\n" +
		"\t tnkinv {subcmd} [params] --token file_with_token \n" +
		"\t   common params: \n" +
		"\t     --account broker|iis|all \n" +
		"\t     --operations filename \n" +
		"\t     --fictives filename \n" +
		"\t     --loglevel {debug|all} \n" +
		"\t   subcmds: \n" +
		"\t     show   [--at 1922/12/28 (default: today)] \n" +
		"\t     story  [--start 1901/01/01 (default: year ago)] \n" +
		"\t            [--period day|week|month (default: month)] \n" +
		"\t            [--format human|table (default: human)] \n" +
		"\t     deals  [--start 1901/01/01 (default: none)] \n" +
		"\t            [--end 1902/02/02 (default: now)] \n" +
		"\t            [--period day|week|month|all (default: month)] \n" +
		"\t     price  --tickers ticker1,ticker2,.. \n" +
		"\t            [--start 1901/01/01 (default: year ago)] \n" +
		"\t            [--end 1902/02/02 (default: now)] \n" +
		"\t     sandbox \n")
}

func getAccountIds(c *client.MyClient, accType string) (accIds []string) {
	if accType == "broker" {
		accIds = append(accIds, "")
		return
	}

	for _, acc := range c.RequestAccounts().Payload.Accounts {
		if accType == "all" || acc.BrokerAccountType == "TinkoffIis" {
			accIds = append(accIds, acc.BrokerAccountID)
		}
	}
	return
}

func main() {
	cmd, cfg := parseCmdline()

	if cfg.token == "" {
		usage()
		log.Fatal("no token provided")
	}

	c := client.NewClient(cfg.token)

	if cmd == "sandbox" {
		c.TrySandbox()
		c.Stop()
		return
	}

	if cmd == "price" {
		portfolio.GetPrices(c, cfg.tickers, cfg.start, cfg.end, cfg.period, cfg.format)
		return
	}

	port := portfolio.NewPortfolio(c, getAccountIds(c, cfg.acc), cfg.sideOps, cfg.fictOps)

	if cmd == "show" {
		port.Collect(cfg.at)
		port.Print(cfg.at)
		return
	}

	if cmd == "deals" {
		if cfg.startSet {
			port.ListDeals(cfg.start, cfg.end)
			return
		}

		since := time.Now()

		if cfg.period == "" {
			cfg.period = "month"
		}

		if cfg.period == "day" {
			since = since.AddDate(0, 0, -1)
		}
		if cfg.period == "week" {
			since = since.AddDate(0, 0, -7)
		}
		if cfg.period == "month" {
			since = since.AddDate(0, -1, 0)
		}
		if cfg.period == "all" {
			since = time.Time{}
		}

		port.ListDeals(since, cfg.end)
		return
	}

	if cmd == "story" {
		if cfg.period == "" {
			cfg.period = "month"
		}

		port.ListBalances(cfg.start, cfg.period, cfg.format)
		return
	}
}
