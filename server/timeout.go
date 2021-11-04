package main

import (
	"errors"
	"net"
	"sync"
	"time"
)

type multipacket_timeout_token struct {
	Pack    chan *packet
	Mut     *sync.Mutex
	Timeout time.Time
}

type packet_timeout_token struct {
	Pack    chan packet
	Mut     sync.Mutex
	Timeout time.Time
}

type conf_token struct{}

//multipacket_timeout is intended to be run in an asynchronous goroutine.
func (t *multipacket_timeout_token) multipacket_timeout(error_channel chan error) {
	var now time.Time = time.Now()
	for t.Timeout.After(now) {
		continue
	}
	for len(t.Pack) > 0 {
		go func(token multipacket_timeout_token) {
			if token.Timeout.Before(time.Now()) {
				token.Mut.Lock()
				close(token.Pack)
				token.Mut.Unlock()
			} else {
				error_channel <- errors.New(ERRMSG_TIMEOUT)
			}
		}(*t)
	}
}

func DefaultSendTimeoutBehaviour(c net.Conn) {
	c.Write([]byte(TIMEOUT_ERROR))
	c.Close()
}

func DefaultTimeout() (t time.Time) {
	t = time.Now().Add(5 * time.Minute)
	return
}
