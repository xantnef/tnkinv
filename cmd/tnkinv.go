package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token  string
	period string
	start  time.Time
}

func parseCmdline() (string, config) {
	if len(os.Args) < 2 {
		usage()
		log.Fatal("no cmd provided")
	}

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

	fs := flag.NewFlagSet("", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	period := fs.String("period", "month", "story period")
	start := fs.String("start", "2018/09/01", "starting point in time (format: 1922/12/28)")
	fs.Parse(os.Args[2:])

	startParsed, err := time.Parse("2006/01/02", *start)
	if err != nil {
		usage()
		log.Fatalf("unrecognized date %s", *start)
	}

	return cmd, config{
		token:  *token,
		period: *period,
		start:  startParsed,
	}
}

func usage() {
	fmt.Printf("usage:\n" +
		"\t tnkinv {show|story|deals} --token file_with_token [--period week|month|year]\n")
}

func main() {
	cmd, cfg := parseCmdline()

	if cfg.token == "" {
		usage()
		log.Fatal("no token provided")
	}

	if cmd == "sandbox" {
		c := client.NewClient(cfg.token)
		c.TrySandbox()
		c.Stop()
		return
	}

	port := portfolio.NewPortfolio(client.NewClient(cfg.token))

	if cmd == "show" {
		port.Collect(cfg.start)
		port.Print()
		return
	}

	if cmd == "deals" {
		t := time.Now()

		if cfg.period == "day" {
			t = t.AddDate(0, 0, -1)
		}
		if cfg.period == "week" {
			t = t.AddDate(0, 0, -7)
		}
		if cfg.period == "month" {
			t = t.AddDate(0, -1, 0)
		}
		if cfg.period == "all" {
			t = cfg.start
		}

		port.ListDeals(t)
		return
	}

	if cmd == "story" {
		port.ListBalances(cfg.start, cfg.period)
		return
	}
}
