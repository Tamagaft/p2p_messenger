package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Chat struct {
	Id    string
	Name  string
	Users int
}

func NewChat(name string) (Chat, error) {
	id, err := randomId()
	if err != nil {
		return Chat{}, err
	}
	return Chat{Id: id, Name: name, Users: 0}, nil
}

func randomId() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (c *Chat) GenerateNewId() {
	id, _ := randomId()
	c.Id = id
}

func (c Chat) MarshalChat() (string, error) {
	chatJson, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error marshaling chat in %s\n%s", chatJson, err)

	}
	return string(chatJson), nil
}
