package main

import (
	"fmt"
	"strings"
)

type Address struct {
	IPv4 string
	Port string
}

func makeAddress(address string) Address {
	ipPort := strings.Split(address, ":")
	ret := Address{
		IPv4: ipPort[0],
		Port: ipPort[1],
	}
	return ret
}

func (address *Address) Get() string {
	return fmt.Sprintf("%s:%s", address.IPv4, address.Port)
}
