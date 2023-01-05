package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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
		ret += fmt.Sprintf("%d: %s\n", i, server.GetString())
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
		ret += fmt.Sprintf("%d: %s: %s\n", i, server.GetString(), chat.Name)
		i++
	}
	return ret
}

func (node *Node) AlterConnectionsChat(nodeIp Address, chat Chat) {
	node.Connections[nodeIp] = chat
}

func (node *Node) ConnectTo(addresses []string) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	for _, e := range addresses {
		msg.Headers.To = makeAddress(e)
		msg.Headers.Type = MESSAGE_TYPE_CONNECTION
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

func (node *Node) ResendToChat(msg *Message, fromIp Address) {
	var wg sync.WaitGroup
	for e, c := range node.Connections {
		if e == fromIp {
			continue
		}
		if c.Name != node.ActiveChat.Name {
			continue
		}
		msg.Headers.To = e
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
		if c.Name != node.ActiveChat.Name {
			continue
		}
		msg := InitMessage(node)
		msg.Headers.To = i
		msg.Headers.Type = MESSAGE_TYPE_TEXT
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

func (node *Node) RemoveConnection(ip Address) {
	delete(node.Connections, ip)
}

func (node *Node) SendChat() {
	var wg sync.WaitGroup
	for address := range node.Connections {
		msg := InitMessage(node)
		msg.Headers.To = address
		msg.Headers.Type = MESSAGE_TYPE_CHAT_CHANGE

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

func (node *Node) DisconnectionMessages() {
	var wg sync.WaitGroup
	msg := InitMessage(node)

	for ip := range node.Connections {
		msg.Headers.To = ip
		msg.Headers.Type = MESSAGE_TYPE_STOP_CONNECTION
		msg.Headers.From = node.Address

		wg.Add(1)
		msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) SendChatTo(ip Address) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	msg.Headers.To = ip
	msg.Headers.Type = MESSAGE_TYPE_CHAT_CHANGE
	msg.Headers.From = node.Address

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
	listener, err := net.Listen("tcp", node.Address.GetString())
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

	msgOriginIp := msg.Headers.From

	switch msgType {
	case MESSAGE_TYPE_TEXT:
		if !messageStore.Exists(&msg) {
			messageStore.Add(&msg)
			node.ResendToChat(&msg, msgOriginIp)
			ShowMessage(msg.Data)
		}
	case MESSAGE_TYPE_CONNECTION:
		var chat Chat
		a := msgOriginIp
		err := json.Unmarshal([]byte(msg.Data), &chat)
		if err != nil {
			log.Printf("Unable to unmarshal chat %s", msg)
		}
		log.Printf("got connection from %s, saved as %s", msg.Headers.From, msgOriginIp)
		node.Connections[a] = chat
		node.SendChatTo(msg.Headers.From)
	case MESSAGE_TYPE_CHAT_CHANGE:
		var newConnectionChat Chat
		err = json.Unmarshal([]byte(msg.Data), &newConnectionChat)
		if err != nil {
			log.Printf("Error: Unable to unmarshal chat in response %s from %s\n%s", message, msg.Headers.From, err)
		}
		node.AlterConnectionsChat(msgOriginIp, newConnectionChat)
	case MESSAGE_TYPE_STOP_CONNECTION:
		node.RemoveConnection(msgOriginIp)
	}
}
