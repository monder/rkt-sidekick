package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"

	"github.com/monder/rkt-sidekick/modes"
)

func main() {
	c := cli.NewCLI("rkt-sidekick", "0.1.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"etcd": modes.EtcdCommand,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
