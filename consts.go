package main

import "time"

const (
	MESSAGE_TYPE_TEXT             = iota
	MESSAGE_TYPE_CONNECTION       = iota
	MESSAGE_TYPE_FIRST_CONNECTION = iota
	MESSAGE_TYPE_CHAT_CHANGE      = iota
	MESSAGE_TYPE_CHAT_CREATION    = iota
	MESSAGE_TYPE_CHAT_ALL         = iota
	MESSAGE_TYPE_STOP_CONNECTION  = iota

	MESSAGE_AUTDATED = time.Minute * 5
)
