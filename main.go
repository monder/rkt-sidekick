package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"

	"github.com/monder/rkt-sidekick/modes"
)

func main() {
	c := cli.NewCLI("rkt-sidekick", "0.1.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"etcd":    modes.EtcdCommand,
		"route53": modes.Route53Command,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
