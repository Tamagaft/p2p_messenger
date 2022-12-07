package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
)

type Node struct {
	Connections map[Address]Chat
	Address     Address
	ActiveChat  Chat
}

func InitNode(address []string) *Node {
	return &Node{Connections: make(map[Address]Chat),
		Address: Address{
			IPv4: address[0],
			Port: address[1]}}
}

func (node *Node) Run(handleServer func(node *Node), client fyne.Window) {
	go handleServer(node)
	client.ShowAndRun()
}

func (node *Node) GetConnections() string {
	var ret string
	var i int
	if len(node.Connections) == 0 {
		return "no connections"
	}
	for server := range node.Connections {
		ret += fmt.Sprintf("%d: %s\n", i, server.Get())
		i++
	}
	return ret
}

func (node *Node) GetConnectionsChat() string {
	var ret string
	var i int
	if len(node.Connections) == 0 {
		return "no connections"
	}
	for server, chat := range node.Connections {
		ret += fmt.Sprintf("%d: %s: %s\n", i, server.Get(), chat.Name)
		i++
	}
	return ret
}

func (node *Node) AlterConnectionsChat(ip string, chat Chat) {
	nodeAddress := makeAddress(ip)
	node.Connections[nodeAddress] = chat
}

func (node *Node) ConnectTo(addresses []string) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	for _, e := range addresses {
		msg.Headers["to"] = e
		msg.Headers["type"] = strconv.Itoa(int(MESSAGE_TYPE_CONNECTION))
		chatJson, err := node.MarshalChat()
		if err != nil {
			log.Print(err.Error())
		}
		msg.Data = chatJson
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) MarshalChat() (string, error) {
	chatJson, err := json.Marshal(node.ActiveChat)
	if err != nil {
		return "", fmt.Errorf("error marshaling chat in SendChat: %s\n%s", chatJson, err)

	}
	return string(chatJson), nil
}

func (node *Node) ResendToChat(msg *Message, fromIp string) {
	var wg sync.WaitGroup
	for e, c := range node.Connections {
		if e == makeAddress(fromIp) {
			continue
		}
		if c != node.ActiveChat {
			continue
		}
		msg.Headers["to"] = e.Get()
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) SendMessageToChat(userInput string) {
	var wg sync.WaitGroup
	if len(node.Connections) == 0 {
		ShowMessage("no connections")
		return
	}
	if node.ActiveChat.Name == "" {
		ShowMessage("no chat")
		return
	}
	for i, c := range node.Connections {
		if c != node.ActiveChat {
			continue
		}
		msg := InitMessage(node)
		msg.Headers["to"] = i.Get()
		msg.Headers["type"] = strconv.Itoa(int(MESSAGE_TYPE_TEXT))
		msg.Data = userInput
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) JoinChat(chat []string) {
	newChat := chat[0]
	node.ActiveChat = Chat{Name: newChat}
	node.SendChat()
}

func (node *Node) SendChat() {
	var wg sync.WaitGroup
	for address := range node.Connections {
		msg := InitMessage(node)
		msg.Headers["to"] = address.Get()
		msg.Headers["type"] = strconv.Itoa(MESSAGE_TYPE_CHAT_CHANGE)

		chatJson, err := node.MarshalChat()
		if err != nil {
			log.Print(err.Error())
			continue
		}

		msg.Data = chatJson

		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) SendChatTo(ip string) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	msg.Headers["to"] = ip
	msg.Headers["type"] = strconv.Itoa(int(MESSAGE_TYPE_CHAT_CHANGE))
	msg.Headers["from"] = node.Address.Get()

	chatJson, err := node.MarshalChat()
	if err != nil {
		log.Print(err.Error())
	}

	msg.Data = chatJson

	wg.Add(1)
	go msg.Send(&wg)
	wg.Wait()
}
func handleServer(node *Node) {
	listener, err := net.Listen("tcp", node.Address.Get())
	if err != nil {
		log.Fatalf("Failed to start listener on address %s:%s", node.Address.IPv4, node.Address.Port)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go handleConnection(node, conn)
	}
}

func handleConnection(node *Node, conn net.Conn) {
	defer conn.Close()
	var (
		buf     = make([]byte, 512)
		message string
		msg     Message
	)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		message += string(buf[:n])
	}
	err := json.Unmarshal([]byte(message), &msg)
	if err != nil {
		log.Printf("Unable to unmarshal %s", message)
	}

	log.Printf("message received:%s", msg)

	msgType, err := msg.GetType()
	if err != nil {
		return
	}

	msgOriginIp := msg.Headers["from"]

	switch msgType {
	case MESSAGE_TYPE_TEXT:
		if !messageStore.Exists(&msg) {
			messageStore.Add(&msg)
			node.ResendToChat(&msg, msgOriginIp)
			ShowMessage(msg.Data)
		}
	case MESSAGE_TYPE_CONNECTION:
		var chat Chat
		a := makeAddress(msgOriginIp)
		err := json.Unmarshal([]byte(msg.Data), &chat)
		if err != nil {
			log.Printf("Unable to unmarshal chat %s", msg)
		}
		log.Printf("got connection from %s, saved as %s", msg.Headers["from"], msgOriginIp)
		node.Connections[a] = chat
		node.SendChatTo(msg.Headers["from"])
	case MESSAGE_TYPE_CHAT_CHANGE:
		var newConnectionChat Chat
		err = json.Unmarshal([]byte(msg.Data), &newConnectionChat)
		if err != nil {
			log.Printf("Error: Unable to unmarshal response %s from %s\n%s", message, msg.Headers["from"], err)
		}
		node.AlterConnectionsChat(msgOriginIp, newConnectionChat)
	}

}
