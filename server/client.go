package main

import (
	"net"
	"time"
)

const (
	SIGN_IN string = "sign_in"
)

type client struct {
	id               int
	send_function    func(i interface{}, c net.Conn, timeout time.Time, error_channel chan error) error
	receive_function func(i interface{}, c net.Conn, timeout time.Time, error_channel chan error) error
	info             map[string]interface{}
}

var native_client client = client{
	id:               nativegoclient,
	send_function:    SendStruct,
	receive_function: ReceiveStruct,
	info: map[string]interface{}{
		SIGN_IN: "sic_native",
	},
}

var js_client client = client{
	id:               jsclient,
	send_function:    SendStruct_JSClient,
	receive_function: ReceiveStruct_JSClient,
}

var c_client client = client{
	id:               cclient,
	send_function:    SendStruct,
	receive_function: ReceiveStruct,
}

func getClient(id int) *client {
	switch id {
	case nativegoclient:
		return &native_client
	case jsclient:
		return &js_client
	case cclient:
		return &c_client
	default:
		return nil
	}
}

type pair struct {
	A interface{}
	B interface{}
}
