package main

import (
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
				node.DisconnectionMessages()
				os.Exit(0)
			case "/cs":
				node.ConnectTo(splited[1:])
			case "/jc":
				node.JoinChat(splited[1:2])
			case "/c":
				log.Print(node.GetConnectionsChat())
				ShowMessage(node.GetConnections())
			case "/a":
				ShowMessage(node.Address.Get())
			case "/g":
				ShowMessage(node.Address.Get())
			case "/h":
				ShowMessage("/exit\n/cs\tconnect to server\n/jc\tjoin-chat\n/c\tshow connections\n/g\tshow address\n/help")
			}
		} else {
			text := userInput
			ShowMessage(text)
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
