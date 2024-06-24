package config

import (
	"strconv"
	"strings"
)

const (
	MinPortNumber = 0
	MaxPortNumber = 65535
)

type NetAddress struct {
	Host string
	Port int
}

func (a NetAddress) Type() string {
	return "NetAddress"
}

func (a NetAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return ErrIncorrectNetAddress
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil || port < MinPortNumber || port > MaxPortNumber {
		return &ErrIncorrectPort{Port: hp[1]}
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}
