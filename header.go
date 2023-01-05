package main

import "time"

type Header struct {
	From Address
	To   Address
	Type int
	Time time.Time
}
