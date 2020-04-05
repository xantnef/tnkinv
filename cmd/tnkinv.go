package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token, sideOps, period, format, acc string

	start, at time.Time

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
	cmds := map[string]bool{
		"sandbox": true,
		"show":    true,
		"story":   true,
		"deals":   true,
	}

	if !cmds[cmd] {
		usage()
		log.Fatalf("unknown command %s", cmd)
	}

	// ------------
	// List options

	fs := flag.NewFlagSet("", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	sideOps := fs.String("operations", "", "json file with operations")
	format := fs.String("format", "human", "output format")
	acc := fs.String("account", "broker", "account")
	loglevel := fs.String("loglevel", "none", "log level")

	period := fs.String("period", "month", "story period")
	start := fs.String("start", "", "starting point in time (format: 1922/12/28; default: year ago)")
	atTime := fs.String("at", "", "point in time (default: now). Not supported yet")

	fs.Parse(os.Args[2:])

	cfg.token = *token
	cfg.sideOps = *sideOps
	cfg.period = *period

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

	formats := map[string]bool{
		"human": true,
		"table": true,
	}
	if !formats[*format] {
		log.Fatalf("bad format %s", *format)
	}
	cfg.format = *format

	// --------------
	// Verify account

	accs := map[string]bool{
		"broker": true,
		"iis":    true,
		"all":    true,
	}
	if !accs[*acc] {
		log.Fatalf("bad account type %s", *acc)
	}
	cfg.acc = *acc

	// ----------------------
	// Parse and verify times

	cfg.start, cfg.startSet = parseDate(*start, time.Now().AddDate(-1, 0, 0))
	cfg.at, _ = parseDate(*atTime, time.Now())

	return cmd, cfg
}

func usage() {
	fmt.Printf("usage:\n" +
		"\t tnkinv {subcmd} [params] --token file_with_token \n" +
		"\t   common params: \n" +
		"\t     --account broker|iis|all \n" +
		"\t     --operations filename \n" +
		"\t     --format human|table \n" +
		"\t     --loglevel {debug|all} \n" +
		"\t   subcmds: \n" +
		"\t     show   [--at 1922/12/28 (default: today)] \n" +
		"\t     story  [--start 1901/01/01 (default: year ago)] \n" +
		"\t            [--period day|week|month (default: month)] \n" +
		"\t     deals  [--start 1901/01/01 (default: none)] \n" +
		"\t            [--period day|week|month|all (default: month)] \n" +
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

	port := portfolio.NewPortfolio(c, getAccountIds(c, cfg.acc), cfg.sideOps)

	if cmd == "show" {
		port.Collect(cfg.at)
		port.Print()
		return
	}

	if cmd == "deals" {
		if cfg.startSet {
			port.ListDeals(cfg.start)
			return
		}

		since := time.Now()

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

		port.ListDeals(since)
		return
	}

	if cmd == "story" {
		port.ListBalances(cfg.start, cfg.period, cfg.format)
		return
	}
}
