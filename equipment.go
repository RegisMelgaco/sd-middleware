package sdmiddleware

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/textproto"
	"time"
)

type (
	Measurement float64
)
type EquipmentName string

type Equipment struct {
	Name EquipmentName
	Sensor
	Interval time.Duration
}

type Sensor struct {
	Monitor
	Max Measurement
	Min Measurement
}

func (e Equipment) Run() {
	for {
		time.Sleep(e.Interval)

		m := e.Read()

		e.Avaluate(m)
	}
}

func (s Sensor) Read() Measurement {
	readingRange := s.Max - s.Min
	r := Measurement(rand.Float64())

	return (r * readingRange) + s.Min
}

type Monitor struct {
	Broker BrokerClient
	Max    Measurement
	Min    Measurement
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
		m.Broker.Send(msg)
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

func (bc BrokerClient) Send(msg string) {
	slog.Info("sending message do broker", slog.String("msg", msg))

	conn, err := net.Dial("tcp", bc.BrokerAddr)
	if err != nil {
		slog.Error("failed to create connection to broker", slog.String("err", err.Error()))

		panic("conn err not implemented")
	}

	err = textproto.NewConn(conn).PrintfLine("%s||%s", bc.Equipment, msg)
	if err != nil {
		panic(fmt.Errorf("failed to write on connection: %w", err))
	}
}
