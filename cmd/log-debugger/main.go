package main

import (
	"fmt"
	"math"
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	log, err := debugger.LoadLog(os.Args[1])
	if err != nil {
		return errors.Wrap(err, "failed to load the log")
	}

	g := debugger.NewGroups()
	for _, entry := range log {
		if err := g.Add(entry); err != nil {
			return errors.Wrap(err, "error adding an entry")
		}
	}

	print(g)

	return nil
}

func print(g *debugger.Groups) {
	for peer, sessions := range g.Peers {
		fmt.Println("peer", peer)

		for _, session := range sessions {
			if session.InititatedBy == debugger.InitiatedByLocalNode {
				fmt.Print("initiated by local ")
			} else {
				fmt.Print("initiated by remote ")
			}
			fmt.Println(math.Abs(float64(session.Number)))

			for _, message := range session.Messages {
				if message.Type == debugger.MessageTypeSent {
					fmt.Print("-> ")
				} else {
					fmt.Print("<- ")
				}

				fmt.Println(
					message.Flags,
					message.RequestNumber,
					message.Body,
				)
			}

			fmt.Println()
		}
	}
}
