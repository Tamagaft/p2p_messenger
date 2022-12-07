package main

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	Headers map[string]string
	Data    string
}

func InitMessage(node *Node) Message {
	headers := make(map[string]string)
	headers["from"] = node.Address.Get()
	msg := Message{Headers: headers}
	currentTime := time.Now()
	msg.Headers["time"] = strconv.Itoa(int(currentTime.UnixMicro()))
	return msg
}

func (msg *Message) Send(wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", msg.Headers["to"])
	if err != nil {
		log.Printf("Error: can not connect to %s\n", msg.Headers["to"])
	}
	defer conn.Close()

	json_msg, err := json.Marshal(*msg)
	if err != nil {
		log.Println("Error: can't marshal message", msg)
	}
	log.Printf("message sent: %s", msg)
	conn.Write(json_msg)
}

func (msg *Message) GetType() (int, error) {
	msgType, ok := msg.Headers["type"]
	if !ok {
		return 0, fmt.Errorf("no type")
	}
	retType, _ := strconv.Atoi(msgType)
	return retType, nil
}

func (m *Message) Hash() string {
	hasher := sha512.New()
	hasher.Write([]byte(m.Headers["from"] + "=>" + m.Headers["to"] + " at " + m.Headers["time"] + " text: " + m.Data))
	return string(hasher.Sum(nil))
}

func (m *Message) Outdated() bool {
	now := time.Now()
	period := MESSAGE_AUTDATED

	unixMicroTime, err := strconv.Atoi(m.Headers["time"])
	if err != nil {
		log.Printf("can not get time from message: %s\n", m)
	}

	t := time.UnixMicro(int64(unixMicroTime))

	return t.Add(period).Before(now)
}
