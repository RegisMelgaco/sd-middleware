package main

import (
	"flag"
	"local/sdmiddleware"
	"time"
)

var (
	equipName     = flag.String("name", "equipment_1", "")
	monitorMax    = flag.Float64("alert-ceiling", 2, "")
	monitorMin    = flag.Float64("alert-floor", 1, "")
	sensorMax     = flag.Float64("sensor-ceiling", 3, "")
	sensorMin     = flag.Float64("sensor-floor", 0, "")
	equipInterval = flag.Duration("interval", time.Second, "")
	brokerAddr    = flag.String("broker", "0.0.0.0:3000", "")
)

func main() {
	flag.Parse()

	eq := sdmiddleware.Equipment{
		Monitor: sdmiddleware.Monitor{
			Name: sdmiddleware.EquipmentName(*equipName),
			Broker: sdmiddleware.BrokerClient{
				BrokerAddr: *brokerAddr,
			},
			Sensor: sdmiddleware.Sensor{
				Max: sdmiddleware.Measurement(*monitorMax),
				Min: sdmiddleware.Measurement(*monitorMin),
			},
		},
		Interval: *equipInterval,
	}

	eq.Run()
}
