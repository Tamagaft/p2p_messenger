package main

import (
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
