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
	Connections map[Address]string
	Address     Address
	ActiveChat  string
	AllChats    map[string]Chat
	Nickname    string
}

func InitNode(address []string) *Node {
	return &Node{Connections: make(map[Address]string),
		Address: Address{
			IPv4: address[0],
			Port: address[1]},
		AllChats: make(map[string]Chat),
	}
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
	for server, chatId := range node.Connections {
		ret += fmt.Sprintf("%d: %s: %s\n", i, server.GetString(), node.AllChats[chatId])
		i++
	}
	return ret
}

func (node *Node) AlterConnectionsChat(nodeIp Address, chatId string) {
	node.Connections[nodeIp] = chatId
}

func (node *Node) SendNewChat(chat Chat) {
	var wg sync.WaitGroup
	fmt.Println(chat)
	msg := InitMessage(node)
	msg.Headers.Type = MESSAGE_TYPE_CHAT_CREATION
	for adr := range node.Connections {
		msg.Headers.To = adr
		chatJson, err := json.Marshal(chat)
		if err != nil {
			log.Print(err.Error())
		}
		msg.Data = string(chatJson)
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) ConnectTo(addresses []string) {
	var firstMessage bool
	if len(node.Connections) == 0 {
		firstMessage = true
	}
	var wg sync.WaitGroup
	msg := InitMessage(node)
	for _, e := range addresses {
		msg.Headers.Type = MESSAGE_TYPE_CONNECTION
		if firstMessage {
			msg.Headers.Type = MESSAGE_TYPE_FIRST_CONNECTION
		}
		msg.Headers.To = makeAddress(e)

		msg.Data = node.ActiveChat
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) ResendToChat(msg *Message) {
	var wg sync.WaitGroup
	for e, c := range node.Connections {
		if e == msg.Headers.From {
			continue
		}
		if c != node.ActiveChat {
			continue
		}
		msg.Headers.To = e
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) ResendToAll(msg *Message) {
	var wg sync.WaitGroup
	for adr := range node.Connections {
		if adr == msg.Headers.From {
			continue
		}
		msg.Headers.To = adr
		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) AddUserToChat(chatId string) {
	if chat, ok := node.AllChats[chatId]; ok {
		chat.Users += 1
		node.AllChats[chatId] = chat
	}
}

func (node *Node) RemoveUserFromChat(chatId string) {
	if chat, ok := node.AllChats[chatId]; ok {
		chat.Users -= 1
		node.AllChats[chatId] = chat
	}
}

func (node *Node) SendMessageToChat(userInput string) {
	var wg sync.WaitGroup
	if len(node.Connections) == 0 {
		ShowMessage("no connections")
		return
	}
	if node.ActiveChat == "" {
		ShowMessage("no chat")
		return
	}
	for i, c := range node.Connections {
		if c == node.ActiveChat {
			msg := InitMessage(node)
			msg.Headers.To = i
			msg.Headers.Type = MESSAGE_TYPE_TEXT
			msg.Headers.Nickname = node.Nickname
			msg.Data = userInput
			wg.Add(1)
			go msg.Send(&wg)
		}
	}
	wg.Wait()
}

func (node *Node) GetAllChats() string {
	var ret string
	for _, chat := range node.AllChats {
		ret += fmt.Sprintf("id:%s, name:%s, users:%d\n", chat.Id, chat.Name, chat.Users)
	}
	log.Println(ret)
	return ret
}

func (node *Node) JoinChat(chatid string) bool {
	if chat, ok := node.AllChats[chatid]; ok {
		node.ActiveChat = chat.Id
		node.AddUserToChat(node.ActiveChat)
		node.SendJoinChat()
		return true
	}
	return false
}

func (node *Node) RemoveConnection(ip Address) {
	delete(node.Connections, ip)
}

func (node *Node) SendJoinChat() {
	var wg sync.WaitGroup
	for address := range node.Connections {
		msg := InitMessage(node)
		msg.Headers.To = address
		msg.Headers.Type = MESSAGE_TYPE_CHAT_CHANGE
		msg.Headers.Nickname = node.Nickname

		msg.Data = node.ActiveChat

		wg.Add(1)
		go msg.Send(&wg)
	}
	wg.Wait()
}

func (node *Node) SendDisconnectionMessages() {
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

func (node *Node) SendAllChatsTo(ip Address) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	msg.Headers.To = ip
	msg.Headers.Type = MESSAGE_TYPE_CHAT_ALL
	msg.Headers.From = node.Address
	chatsJson, err := json.Marshal(node.AllChats)
	if err != nil {
		log.Println("error marshaling chat in: %s\n%s", chatsJson, err)
		return
	}
	msg.Data = string(chatsJson)
	wg.Add(1)
	go msg.Send(&wg)
	wg.Wait()
}

func (node *Node) SendChatTo(ip Address) {
	var wg sync.WaitGroup
	msg := InitMessage(node)
	msg.Headers.To = ip
	msg.Headers.Type = MESSAGE_TYPE_CHAT_CHANGE
	msg.Headers.From = node.Address

	chatJson, err := node.AllChats[node.ActiveChat].MarshalChat()
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
			node.ResendToChat(&msg)
			ShowNamedMessage(msg.Headers.Nickname, msg.Data)
		}

	case MESSAGE_TYPE_CONNECTION:
		a := msgOriginIp
		log.Printf("got connection from %s, saved as %s", msg.Headers.From, msgOriginIp)
		node.Connections[a] = msg.Data
		node.SendChatTo(msg.Headers.From)

	case MESSAGE_TYPE_FIRST_CONNECTION:
		a := msgOriginIp
		log.Printf("got connection from %s, saved as %s", msg.Headers.From, msgOriginIp)
		node.Connections[a] = msg.Data
		node.SendAllChatsTo(msg.Headers.From)
		node.SendChatTo(msg.Headers.From)

	case MESSAGE_TYPE_CHAT_CHANGE:
		node.AlterConnectionsChat(msgOriginIp, msg.Data)
		node.AddUserToChat(msg.Data)

	case MESSAGE_TYPE_CHAT_CREATION:
		var newChat Chat
		if err = json.Unmarshal([]byte(msg.Data), &newChat); err != nil {
			log.Printf("Error: Unable to unmarshal created chat in %s from %s\n%s", message, msg.Headers.From, err)
			break
		}
		node.AllChats[newChat.Id] = newChat
		if !messageStore.Exists(&msg) {
			messageStore.Add(&msg)
			node.ResendToAll(&msg)
		}

	case MESSAGE_TYPE_CHAT_ALL:
		chats := make(map[string]Chat)
		if err = json.Unmarshal([]byte(msg.Data), &chats); err != nil {
			log.Printf("Error: Unable to unmarshal created chat in %s from %s\n%s", message, msg.Headers.From, err)
			break
		}
		node.AllChats = chats

	case MESSAGE_TYPE_STOP_CONNECTION:
		node.RemoveUserFromChat(node.Connections[msgOriginIp])
		node.RemoveConnection(msgOriginIp)
	}
}
