package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

/*Dr. Muller:
Have your Ambassador function take the first message and determine what kind of client you're dealing with.

Method: Capture transient first-message. If something expected, then handle as expected, otherwise, flag it as an exception, log it, leave.

Note: If you need to add a preamble message to a protocol so that we know who we're talking to, then it's not bloat. Just do it.

If you're trying to put a message in a byte string, then try to read the whole message in [size of buffer] bytes. If you don't have all the bytes in the message,
then read the rest of the buffer again.

must determine when to stop reading a message. Must know length or terminal character or something.

General Strategy (Covered in Stevens): how to encapsulate message in byte string
*/

//handles requests for the main
func Multiplexer(conn net.Conn, error_channel chan error) {
	log.Printf("Connection IP address: %v\n", conn.RemoteAddr())
	var msg Netmessage
	var err error
	var client_type int //enum (const)
	var timeout time.Time

	timeout = DefaultTimeout()
	client_type, err = ClientIdentifier(conn)
	if err != nil {
		Errhandle_Log(err, "Error occured in ClientIdentifier function. Terminating connection./n")
		return
	}
	fmt.Fprintf(os.Stderr, "Connection received from device with enum: %d\n", client_type)

	switch client_type {
	case nativegoclient:
		{
			err = ReceiveStruct(&msg, conn, timeout, error_channel)
			Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
			if err != nil {
				return
			}
		}
	case jsclient:
		{
			err = ReceiveStruct_JSClient(&msg, conn, timeout, error_channel)
			if err != nil {
				return
			}
		}
	case cclient:
		{
			var read_buffer []byte

			read_buffer = make([]byte, 5)
			conn.Read(read_buffer)
			// if rdln != 5 || err != nil {
			// 	return
			// }
			fmt.Printf("%s\n", read_buffer)
		}
	}

	fmt.Printf("MESSAGE RECEIVED: %v\n", msg)
	switch msg.Message {
	case NETREQ_NEWDEVICE:
		RegisterDevice(conn, timeout, error_channel)
	case NETREQ_NEWDEVICE_JAVASCRIPT:
		RegisterDevice_Websocket(conn, timeout, error_channel)
	case NETREQ_TRUSTREQUEST:
		Trust_request(conn, timeout, error_channel)
	case NETREQ_NEWCONTENTTRANSFER:
		ContentTransfer(conn, timeout, error_channel)
	default:
		conn.Close()
	}
}

func ClientIdentifier(conn net.Conn) (int, error) {
	var httpreq *http.Request
	var err error
	var bufioReader *bufio.Reader = bufio.NewReader(conn)
	var retval int
	if err != nil {
		return retval, err
	}
	httpreq, err = http.ReadRequest(bufioReader)
	Errhandle_Log(err, ERRMSG_NETWORK_PARSE_HTTPREQ)
	if err != nil {
		return -1, err
	}
	fmt.Printf("Header: %v\n", httpreq.Header)
	if httpreq.Header["Sec-Websocket-Version"] != nil {
		var clienthash string = httpreq.Header["Sec-Websocket-Key"][0]
		clienthash += "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

		//https: //yalantis.com/blog/how-to-build-websockets-in-go/
		hash := sha1.New()
		hash.Write([]byte(clienthash))
		var rethash string = base64.StdEncoding.EncodeToString(hash.Sum(nil))
		retstring := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-Websocket-Accept: %s\r\n\r\n"
		retstring = fmt.Sprintf(retstring, rethash)
		fmt.Printf("\n\rresponse:\n%s\n", retstring)
		conn.Write([]byte(retstring))
		return jsclient, nil
	} else if httpreq.Header["Nativegoclient"] != nil {
		return nativegoclient, nil
	} else if reflect.DeepEqual(httpreq.Header["User-Agent"], []string{"hermes-C-client"}) {
		return cclient, nil
	} else {
		return -1, nil
	}
}

//Request to add/register device.
func RegisterDevice(conn net.Conn, timeout time.Time, error_channel chan error) (ret_device Device, err error) {
	var ip_string string
	var ip net.IP
	var cdinf DeviceInfo
	var msg Netmessage

	timeout = DefaultTimeout()
	msg = NewNetmessage(NETREQ_NEWDEVICE)
	err = SendStruct(&msg, conn, timeout, error_channel)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return ret_device, err
	}
	err = ReceiveStruct(&cdinf, conn, timeout, error_channel)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
	if err != nil {
		return ret_device, err
	}
	log.Printf("native go client: userid: %s, device: %s", cdinf.Userid, cdinf.Devicename)
	ip_string = strings.Split(conn.RemoteAddr().String(), ":")[0]
	ip = net.ParseIP(ip_string)
	ret_device, err = NewDevice(cdinf.Userid, cdinf.Devicename, ip)

	Errhandle_Log(err, ERRMSG_CREATE_DEVICE)
	SendStruct(&ret_device, conn, timeout, error_channel)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if len(error_channel) != 0 {
		return
	}
	//return or return, because we can't return a nil device
	return ret_device, err
}

func RegisterDevice_Websocket(conn net.Conn, timeout time.Time, error_channel chan error) (Device, error) {
	var retdevice Device
	var err error
	var cdinf DeviceInfo
	var ip net.IP
	var ipstring string

	err = ReceiveStruct_JSClient(&cdinf, conn, timeout, error_channel)
	if err != nil {
		return retdevice, err
	}
	ipstring = strings.Split(conn.RemoteAddr().String(), ":")[0]
	ip = net.ParseIP(ipstring)
	retdevice, err = NewDevice(cdinf.Userid, cdinf.Devicename, ip)
	Errhandle_Log(err, ERRMSG_CREATE_DEVICE)
	log.Printf("DEVICE: %v\n", retdevice)
	err = SendStruct_JSClient(&retdevice, conn, timeout, error_channel)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	//return or return, because we can't return a nil device
	return retdevice, err
}

/*
A TrustConsent Request constitutes a request to send an item of content to another user.

1. server	2. recipient
ACK			userid
ACK/REJ (TERM)

SERVER: forward trustconsent to appropriate channel
*/
func Trust_request(conn net.Conn, timeout time.Time, error_channel chan error) (err error) {
	var recipient string
	var local_errchann chan error
	msg := Netmessage{Message: NETREQ_TRUSTREQUEST}
	err = SendStruct(msg, conn, timeout, local_errchann)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	err = ReceiveStruct(&recipient, conn, timeout, local_errchann)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)

	if err != nil {
		error_channel <- err
		return err
	}
	//TODO: the below is just temp code. It's only useful for compilation while building trust consent struct.
	//This should be the initial endpoint for all things directly "privacy and security"-related.
	if err != nil {
		error_channel <- err
		return
	}
	err = ReceiveStruct(err, conn, timeout, error_channel)
	return
}

func ContentTransfer(conn net.Conn, timeout time.Time, error_channel chan error) (err error) {
	msg := Netmessage{Message: NETREQ_NEWCONTENTTRANSFER}
	err = SendStruct(msg, conn, timeout, error_channel)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		error_channel <- err
		return err
	}

	if err != nil {
		error_channel <- err
		return
	}
	err = ReceiveStruct(err, conn, timeout, error_channel)
	return

}
