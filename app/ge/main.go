package main

import (
	"flag"
	"local/sdmiddleware"
	"log/slog"
)

var (
	webAddr    = flag.String("web", "0.0.0.0:3002", "")
	brokerAddr = flag.String("broker", "0.0.0.0:3001", "")
)

func main() {
	ge := sdmiddleware.GE{
		Addr: *webAddr,
		Broker: sdmiddleware.BrokerClient{
			BrokerAddr: *brokerAddr,
		},
	}

	slog.Info("starting")

	ge.Run()
}
