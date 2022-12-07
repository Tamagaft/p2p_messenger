package main

import "time"

const (
	MESSAGE_TYPE_TEXT        = 1 << iota
	MESSAGE_TYPE_CONNECTION  = 1 << iota
	MESSAGE_TYPE_CHAT_CHANGE = 1 << iota

	MESSAGE_AUTDATED = time.Minute * 5
)
