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
)

func main() {
	flag.Parse()

	eq := sdmiddleware.Equipment{
		Name: sdmiddleware.EquipmentName(*equipName),
		Sensor: sdmiddleware.Sensor{
			Monitor: sdmiddleware.Monitor{
				Max: sdmiddleware.Measurement(*monitorMax),
				Min: sdmiddleware.Measurement(*monitorMin),
			},
			Max: sdmiddleware.Measurement(*sensorMax),
			Min: sdmiddleware.Measurement(*sensorMin),
		},
		Interval: *equipInterval,
	}
}
