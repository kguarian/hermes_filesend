package main

import (
	"errors"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

/*
	Files consist of sets of packets spanning the interval [0, MAX).
*/
type packet struct {
	Cinf  *contentinfo
	Key   *uuid.UUID
	Lower int
	Upper int
	Buf   []byte
}

func send_content(c chan *packet, d Device, conn net.Conn, conn_mut *sync.Mutex, file_sz int, timeout_duration time.Duration, error_channel chan error) {
	var start_time, timeout_time time.Time
	var timeout_token multipacket_timeout_token
	var local_errchann chan error = make(chan error)
	var confirmation_counter int
	var packet_count int
	var currClient *client
	var currpacket *packet
	var err error

	//WE must send ~something~
	if len(c) == 0 {
		error_channel <- errors.New(ERRMSG_EMPTY_TRANSMISSION)
		return
	}

	currClient = getClient(d.DeviceType)

	//set timeout
	start_time = time.Now()
	timeout_time = start_time.Add(timeout_duration)
	timeout_token = multipacket_timeout_token{Pack: c, Mut: &sync.Mutex{}, Timeout: timeout_time}
	timeout_token.multipacket_timeout(error_channel)

	packet_count = len(c)
	//ensure that the connection is ours, then proceed
	conn_mut.Lock()

	//communicate to client quantity of packets to be sent
	Send_with_confirmation(d.DeviceType, &confirmation_counter, &packet_count, conn, timeout_time, local_errchann)
	confirmation_counter = 0

	//begin timeout routine
	go timeout_token.multipacket_timeout(error_channel)

	//send all packets
	for i := 0; i < packet_count; i++ {
		currpacket = <-timeout_token.Pack
		go Send_with_confirmation(d.DeviceType, &confirmation_counter, currpacket, conn, timeout_time, local_errchann)
	}
	//wait for completion
	for confirmation_counter != packet_count && timeout_time.After(time.Now()) {
		continue
	}

	//clean up
	conn_mut.Unlock()
	if timeout_time.After(time.Now()) && confirmation_counter < packet_count {
		message := NewNetmessage(TIMEOUT_ERROR)
		currClient.send_function(&message, conn, DefaultTimeout(), error_channel)
		conn.Close()
		return
	}

	//error flush
	for len(error_channel) != 0 {
		err = <-error_channel
		//untested, this is a reminder.
	}
	Errhandle_Exit(err, err.Error())
}

func receive_content(d Device, c net.Conn, conn_mut *sync.Mutex, file_sz int, timeout_duration time.Duration, error_channel chan error) {
	var packets chan *packet
	var packet_bank []*packet
	var packet_count int
	var currClient *client
	var local_error_channel chan error
	var timeout time.Time
	var multipacket_timeout multipacket_timeout_token
	var content []*packet

	local_error_channel = make(chan error)
	currClient = getClient(d.DeviceType)

	//set timeout
	timeout = DefaultTimeout()
	multipacket_timeout = multipacket_timeout_token{Pack: packets, Mut: conn_mut, Timeout: timeout}
	multipacket_timeout.multipacket_timeout(error_channel)

	currClient.receive_function(&packet_count, c, timeout, local_error_channel)
	if len(local_error_channel) != 0 {
		error_channel <- <-local_error_channel
		return
	}
	packet_bank = make([]*packet, packet_count)
	go func() {
		for i := 0; i < packet_count; i++ {
			currClient.receive_function(&packet_bank[i], c, timeout, local_error_channel)
			packets <- packet_bank[i]
		}
	}()
	for len(packets) != packet_count {
		if len(local_error_channel) != 0 {
			error_channel <- <-local_error_channel
			return
		}
		continue
	}
	content = make([]*packet, packet_count)
	for i := 0; i < packet_count; i++ {
		content[i] = <-packets
	}
	sort.SliceStable(content, func(i int, j int) bool { return content[i].Lower < content[j].Lower })

	message := NewNetmessage(NETMESS_SUCCESSFUL_TRANSMISSION)

	currClient.send_function(message, c, timeout, error_channel)
}

func Send_with_confirmation(client_type int, confirmation_counter *int, i interface{}, c net.Conn, timeout time.Time, error_channel chan error) {
	getClient(client_type).send_function(i, c, timeout, error_channel)
	(*confirmation_counter)++
}
