package sdmiddleware

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

type (
	Measurement float64
)
type EquipmentName string

type Equipment struct {
	Monitor
	Interval time.Duration
}

type Sensor struct {
	Max Measurement
	Min Measurement
}

func (e Equipment) Run() (stop func(bool)) {
	isStopped := false
	stop = func(s bool) {
		isStopped = s
	}

	go func() {
		for {
			if !isStopped {
				m := e.Read()

				e.Avaluate(m)

				time.Sleep(e.Interval)
			}
		}
	}()

	return stop
}

func (s Sensor) Read() Measurement {
	readingRange := s.Max - s.Min
	r := Measurement(rand.Float64())

	return (r * readingRange) + s.Min
}

type Monitor struct {
	Sensor
	Broker BrokerClient
	Max    Measurement
	Min    Measurement
	Name   EquipmentName
}

func (m Monitor) Avaluate(v Measurement) {
	slog.Info("Monitor is evaluating", slog.Float64("measument", float64(v)))

	var err error
	if v >= m.Max {
		err = ErrMaxMeasument
	}

	if v <= m.Min {
		err = ErrMinMeasument
	}

	if err != nil {
		msg := fmt.Sprintf("%s: reading=%v", err.Error(), v)
		m.Broker.Send(string(m.Name), msg)
	}
}

var (
	ErrMaxMeasument = errors.New("measument is higher than allowed range")
	ErrMinMeasument = errors.New("measument is lower than allowed range")
)
