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
	"net/url"
	"strconv"
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
	Msgs []Msg
}

type Msg struct {
	ID    ID
	Value string
}

type ID int

var (
	newIDMux  sync.Mutex
	idCounter int
)

func newID() ID {
	newIDMux.Lock()
	defer newIDMux.Unlock()

	defer func() { idCounter += 1 }()

	return ID(idCounter)
}

func (b *Broker) ListenEquipments() (end chan bool) {
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

			go b.handleNewMsg(conn)
		}
	}()

	return end
}

func (b *Broker) handleNewMsg(c net.Conn) {
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
		topic = Topic{Name: equipment, Msgs: []Msg{}}
	}

	topic.Msgs = append(topic.Msgs, Msg{
		ID:    newID(),
		Value: msg,
	})

	b.topics[equipment] = topic

	slog.Info("broker updated", slog.Any("topics", b.topics))
}

func (b *Broker) ListenGE() (end chan bool) {
	end = make(chan bool)
	go func() {
		defer func() { close(end) }()

		mux := http.NewServeMux()
		mux.HandleFunc("/topics", b.handleListMsgs)
		mux.HandleFunc("DELETE /topics/{id}", b.handleDeleteMsg)

		err := http.ListenAndServe(b.GEAddr, mux)
		if err != nil {
			err = fmt.Errorf("stopped listening ge requests: %w", err)
			panic(err)
		}
	}()

	return end
}

func (b *Broker) handleListMsgs(w http.ResponseWriter, r *http.Request) {
	topics := make([]Topic, 0, len(b.topics))
	for _, t := range b.topics {
		topics = append(topics, t)
	}

	err := json.NewEncoder(w).Encode(topics)
	if err != nil {
		err = fmt.Errorf("marshalling topics: %w", err)
		panic(err)
	}
}

func (b *Broker) handleDeleteMsg(_ http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		err = fmt.Errorf("parse id from path: %w: id=%v", err, r.PathValue("id"))
		panic(err)
	}

	for ti, topic := range b.topics {
		for mi, msg := range topic.Msgs {
			if msg.ID == ID(id) {
				topic.Msgs = append(topic.Msgs[:mi], topic.Msgs[mi+1:]...)
				b.topics[ti] = topic

				slog.Info("deleted message", slog.Any("topics", b.topics))

				return
			}
		}
	}
}

type BrokerClient struct {
	BrokerAddr string
}

func (bc BrokerClient) Send(equipment, msg string) {
	conn, err := net.Dial("tcp", bc.BrokerAddr)
	if err != nil {
		slog.Error("failed to create connection to broker", slog.String("err", err.Error()))

		panic("conn err not implemented")
	}

	defer conn.Close()

	err = textproto.NewConn(conn).PrintfLine("%s||%s", equipment, msg)
	if err != nil {
		panic(fmt.Errorf("failed to write on connection: %w", err))
	}
}

func (bc BrokerClient) List() []Topic {
	url := url.URL{
		Scheme: "http",
		Host:   bc.BrokerAddr,
		Path:   "/topics",
	}

	resp, err := http.Get(url.String())
	if err != nil {
		err := fmt.Errorf("requesting topics: %w", err)
		panic(err)
	}

	var topics []Topic
	err = json.NewDecoder(resp.Body).Decode(&topics)
	if err != nil {
		err = fmt.Errorf("decode topics list: %w", err)
		panic(err)
	}

	return topics
}

func (bc BrokerClient) Delete(id ID) {
	url := url.URL{
		Scheme: "http",
		Host:   bc.BrokerAddr,
		Path:   fmt.Sprintf("/topics/%d", id),
	}

	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		err = fmt.Errorf("creating delete msg request: %w", err)
		panic(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		err := fmt.Errorf("requesting topics: %w", err)
		panic(err)
	}
}
