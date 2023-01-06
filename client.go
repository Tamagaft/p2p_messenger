package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var msgContainer *fyne.Container

func InitClient(msgBlock *fyne.Container) {
	msgContainer = msgBlock
}

func ShowMessage(text string) {
	msgContainer.Add(widget.NewLabel(text))
}

func ShowNamedMessage(name, text string) {
	if name == "" {
		name = "anon"
	}
	msgContainer.Add(widget.NewLabel(fmt.Sprintf("%s:\n %s", name, text)))
}
