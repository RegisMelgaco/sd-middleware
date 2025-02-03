package sdmiddleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/textproto"
	"strings"
	"sync"
)

type Broker struct {
	topics           map[EquipmentName]Topic
	MeasurementsAddr string
	GEAddr           string
	tm               sync.Mutex
}

type Topic struct {
	Name EquipmentName
	Msgs []string
}

func (b *Broker) ListenMeasurements() (end chan bool) {
	end = make(chan bool)

	go func() {
		defer func() { close(end) }()

		lis, err := net.Listen("tcp", b.MeasurementsAddr)
		if err != nil {
			err = fmt.Errorf("listening open conn request for measurements: %w", err)
			panic(err)
		}

		for {
			conn, err := lis.Accept()
			if err != nil {
				err = fmt.Errorf("accept conn: %w", err)
				panic(err)
			}

			go b.handleNewMeasuments(conn)
		}
	}()

	return end
}

func (b *Broker) handleNewMeasuments(c net.Conn) {
	defer c.Close()

	conn := textproto.NewConn(c)

	for {
		var (
			data []byte
			err  error
		)

		for {
			data, err = conn.ReadLineBytes()
			if err != nil && errors.Is(err, io.EOF) {
				continue
			}

			break
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			err = fmt.Errorf("read measument from conn: %w", err)
			panic(err)
		}

		parts := strings.Split(string(data), "||")
		if len(parts) != 2 {
			panic("measument msg with unexpected format")
		}

		b.appendMsg(EquipmentName(parts[0]), parts[1])
	}
}

func (b *Broker) appendMsg(equipment EquipmentName, msg string) {
	b.tm.Lock()
	defer b.tm.Unlock()

	slog.Info("appending msg", slog.String("equipment", string(equipment)), slog.String("msg", msg))

	if b.topics == nil {
		b.topics = make(map[EquipmentName]Topic)
	}

	topic, ok := b.topics[equipment]
	if !ok {
		topic = Topic{Name: equipment, Msgs: []string{}}
	}

	topic.Msgs = append(topic.Msgs, msg)

	b.topics[equipment] = topic

	slog.Info("broker updated", slog.Any("topics", b.topics))
}

func (b *Broker) ListenGE() (end chan bool) {
	end = make(chan bool)
	go func() {
		defer func() { close(end) }()

		mux := http.NewServeMux()
		mux.Handle("GET /measurement", http.HandlerFunc(b.handleListMeasuments))

		err := http.ListenAndServe(b.GEAddr, mux)
		if err != nil {
			err = fmt.Errorf("stopped listening ge requests: %w", err)
			panic(err)
		}
	}()

	return end
}

func (b *Broker) handleListMeasuments(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(b.topics)
	if err != nil {
		err = fmt.Errorf("marshalling measuments: %w", err)
		panic(err)
	}
}
