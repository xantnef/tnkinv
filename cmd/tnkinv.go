package main

import (
	"flag"
	"fmt"
	"os"

	"../pkg/myclient"
)

func parseCmdline() *myclient.Config {
	token := flag.String("token", "", "API token")
	sandToken := flag.String("sandtoken", "", "sandbox API token")

	flag.Parse()

	return &myclient.Config{
		Token:     *token,
		SandToken: *sandToken,
	}
}

func main() {
	c, err := myclient.NewClient(parseCmdline())
	if err != nil {
		fmt.Errorf("cannot create:", err)
		os.Exit(-1)
	}

	err = c.Run()
	if err != nil {
		fmt.Errorf("cannot run:", err)
		os.Exit(-1)
	}
	c.Stop()
}
