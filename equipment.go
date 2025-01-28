package sdmiddleware

import (
	"errors"
	"math/rand"
	"time"
)

type (
	Measurement float32
)
type EquipmentName string

type Equipment struct {
	Sensor
	Name     EquipmentName
	Interval time.Duration
}

type Sensor struct {
	Monitor
	Max Measurement
	Min Measurement
}

func (e Equipment) Run() (end func() error) {
	wait := make(chan bool)
	err := make(chan error, 1)

	go func() {
		for {
			switch {
			case <-wait:
				close(err)
				break
			default:
				time.Sleep(e.Interval)

				m := e.Sensor.Read()
				e.Sensor.Monitor.OnRead(m)
			}
		}
	}()

	return func() error {
		close(wait)
	}
}

func (s Sensor) Read() Measurement {
	readingRange := s.Max - s.Min
	r := Measurement(rand.Float32())

	return (r * readingRange) + s.Min
}

type Monitor struct {
	broker BrokerClient
	Max    Measurement
	Min    Measurement
}

func (m Monitor) OnRead(v Measurement) {
	if v >= m.Max {
		m.broker.Send(ErrMaxMeasument)
	}

	if v <= m.Min {
		m.broker.Send(ErrMinMeasument)
	}
}

var (
	ErrMaxMeasument = errors.New("measument is higher than allowed range")
	ErrMinMeasument = errors.New("measument is lower than allowed range")
)

type BrokerClient struct {
	BrokerAddr string
	Equipment  EquipmentName
}

func (bc BrokerClient) Send(err error) {
	panic("implement me")
}
