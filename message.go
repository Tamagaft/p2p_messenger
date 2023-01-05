package main

import (
	"crypto/sha512"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

type Message struct {
	Headers Header
	Data    string
}

func InitMessage(node *Node) Message {
	headers := Header{}
	headers.From = node.Address
	msg := Message{Headers: headers}
	msg.Headers.Time = time.Now()
	return msg
}

func (msg *Message) Send(wg *sync.WaitGroup) {
	defer wg.Done()

	json_msg, err := json.Marshal(*msg)
	if err != nil {
		log.Println("Error: can't marshal message", msg)
		return
	}

	conn, err := net.Dial("tcp", msg.Headers.To.GetString())
	if err != nil {
		log.Printf("Error: can not connect to %s\n", msg.Headers.To)
		return
	}
	defer conn.Close()

	log.Printf("message sent: %s", msg)
	conn.Write(json_msg)
}

func (msg *Message) GetType() (int, error) {
	msgType := msg.Headers.Type

	return msgType, nil
}

func (m *Message) Hash() string {
	hasher := sha512.New()
	hasher.Write([]byte(m.Headers.From.GetString() + "=>" + m.Headers.To.GetString() + " at " + m.Headers.Time.Format("2006-01-02 15:04:05") + " text: " + m.Data))
	return string(hasher.Sum(nil))
}

func (m *Message) Outdated() bool {
	now := time.Now()
	period := MESSAGE_AUTDATED

	msgTime := m.Headers.Time

	return msgTime.Add(period).Before(now)
}
