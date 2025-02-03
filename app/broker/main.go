package main

import (
	"flag"
	"local/sdmiddleware"
	"log/slog"
)

var (
	geAddr         = flag.String("ge", "0.0.0.0:3001", "")
	equipmentsAddr = flag.String("equipments", "0.0.0.0:3000", "")
)

func main() {
	flag.Parse()

	b := sdmiddleware.Broker{
		MeasurementsAddr: *equipmentsAddr,
		GEAddr:           *geAddr,
	}

	slog.Info("starting")

	endGE := b.ListenGE()
	endE := b.ListenEquipments()

	<-endGE
	<-endE
}
