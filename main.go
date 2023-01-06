package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage %s ip:port", os.Args[0])
	}

	portIpInput := strings.Split(os.Args[1], ":")
	if len(portIpInput) != 2 {
		log.Fatalf("Address should be ip:port")
	}

	InitHasher()

	node := InitNode(portIpInput)

	//-----------------------------------------

	myApp := app.New()
	myWindow := myApp.NewWindow("Entry Widget")
	myWindow.Resize(fyne.NewSize(400, 600))

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	messagesBlock := container.New(layout.NewGridLayout(1))
	InitClient(messagesBlock)
	messagesScrollBlock := container.NewVScroll(messagesBlock)

	textInput := container.NewVBox(input, widget.NewButton("Save", func() {
		userInput := input.Text
		if userInput == "" {
			return
		}
		if userInput[0] == '/' {
			splited := strings.Split(userInput, " ")
			switch splited[0] {
			case "/exit":
				node.SendDisconnectionMessages()
				os.Exit(0)
			case "/cs":
				node.ConnectTo(splited[1:])
			case "/lc":
				ShowMessage(node.GetAllChats())
				for id, e := range node.AllChats {
					fmt.Printf("%s: %s\n", id, e.Users)
				}
			case "/jc":
				if node.JoinChat(splited[1]) {
					ShowMessage(fmt.Sprintf("Joined chat: %s", node.ActiveChat))
				} else {
					ShowMessage(fmt.Sprintf("Chat id %s not found", splited[1]))
				}
			case "/c":
				log.Print(node.GetConnectionsChat())
				ShowMessage(node.GetConnections())
			case "/a":
				ShowMessage(node.Address.GetString())
			case "/g":
				ShowMessage(node.Address.GetString())
			case "/cc":
				chat, err := NewChat(splited[1])
				if err != nil {
					log.Println(err)
					ShowMessage("Chat creation failed。")
				}
				_, ok := node.AllChats[chat.Id]
				for ok {
					chat.GenerateNewId()
					_, ok = node.AllChats[chat.Id]
				}
				node.AllChats[chat.Id] = chat
				ShowMessage(fmt.Sprintf("Chat created:\n%s: %s", chat.Id, chat.Name))
				node.SendNewChat(chat)
			case "/sn":
				node.Nickname = strings.Join(splited[1:], " ")
			case "/h":
				ShowMessage("/exit\n/cs\tconnect to server\n/jc\tjoin-chat\n/c\tshow connections\n/g\tshow address\n/help")
			}
		} else {
			text := userInput
			ShowNamedMessage(node.Nickname, text)
			node.SendMessageToChat(text)
		}
		input.SetText("")
	}))

	content := container.NewVSplit(messagesScrollBlock, textInput)
	content.SetOffset(0.86)

	myWindow.SetContent(content)

	//-----------------------------------------

	node.Run(handleServer, myWindow)

}
