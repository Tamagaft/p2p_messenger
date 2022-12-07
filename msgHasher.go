package main

import (
	"sync"
	"time"
)

type MessageStore struct {
	messages map[string]*Message
	mx       sync.Mutex
}

var messageStore *MessageStore

func InitHasher() {
	messageStore = &MessageStore{
		messages: make(map[string]*Message, 0),
		mx:       sync.Mutex{},
	}

	go func() {
		for {
			messageStore.mx.Lock()
			for hash, message := range messageStore.messages {
				if message.Outdated() {
					delete(messageStore.messages, hash)
				}
			}
			messageStore.mx.Unlock()
			time.Sleep(MESSAGE_AUTDATED)
		}
	}()
}

func (ms *MessageStore) Exists(msg *Message) bool {
	ms.mx.Lock()
	exists := ms.messages[msg.Hash()] != nil
	ms.mx.Unlock()

	return exists
}

func (ms *MessageStore) Add(msg *Message) {
	ms.mx.Lock()
	ms.messages[msg.Hash()] = msg
	ms.mx.Unlock()
}
