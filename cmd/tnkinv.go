package main

import (
	"flag"
	"log"
	"os"

	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token, period string
}

func parseCmdline() (string, config) {

	if len(os.Args) < 2 {
		log.Fatal("no cmd provided")
	}

	cmd := os.Args[1]
	cmds := map[string]bool{
		"sandbox": true,
		"show":    true,
		"story":   true,
	}

	if !cmds[cmd] {
		log.Fatalf("unknown command %s", cmd)
	}

	fs := flag.NewFlagSet("", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	period := fs.String("period", "month", "story period")

	fs.Parse(os.Args[2:])

	return cmd, config{
		token:  *token,
		period: *period,
	}
}

func main() {
	cmd, cfg := parseCmdline()

	if cfg.token == "" {
		log.Fatal("no token provided")
		return
	}

	if cmd == "sandbox" {
		c := client.NewClient(cfg.token)
		c.TrySandbox()
		c.Stop()
		return
	}

	port := portfolio.NewPortfolio(client.NewClient(cfg.token))

	if cmd == "show" {
		port.Collect()
		port.Print()
		return
	}

	if cmd == "story" {
		port.ListBalances(cfg.period)
		return
	}
}
