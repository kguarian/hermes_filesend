package main

//USING BIG-ENDIAN
//REF: 	https://www.ibm.com/docs/en/error?originalUrl=SSB27U_6.4.0/com.ibm.zvm.v640.kiml0/asonetw.htm#:~:text=The%20network%20byte%20order%20is,confusion%20because%20of%20byte%20ordering.
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
)

//Message struct to complement the send/receive struct methods
type Netmessage struct {
	Message string
}

type Websocket_datagram struct {
	Fin      bool
	Opcode   byte
	data_len uint64
	Data     []byte
}

/*
	ReadDatagram will read in a datagram from the connection, returning an error
	if one of its called methods returns an error. If the necessary bytes don't come
	through the connection, this function won't stop blocking... yet.

	I've attempted to implement this function in accordance with
	https://tools.ietf.org/html/rfc6455#section-5.2 (5/3/2021)

*/
func (*Websocket_datagram) ReadDatagram(c net.Conn) (Websocket_datagram, error) {
	var rcv_len int
	var err error
	var buf []byte
	var currbits uint16
	var retDG Websocket_datagram

	/*
		Indicates that this is the final fragment in a message.  The first
		fragment MAY also be the final fragment.
	*/
	var fin bool
	var opcode byte
	var mask bool
	var mask_key []byte
	var data_len uint64
	var data []byte

	buf = make([]uint8, 2)
	rcv_len, err = c.Read(buf)
	if err != nil || rcv_len != len(buf) {
		return retDG, errors.New(ERRMSG_NETWORK_READ + " occurrence 1")
	}
	//loading first 2 bytes into bit frame
	currbits = (uint16(buf[0]) << 8) + uint16(buf[1])
	//fin set if bit 0 is 1, so shift off bits 1-15
	fin = currbits>>15 == 1
	//opcode is bit 4-7, so shift off bits 0-3 and 8-15
	opcode = byte((currbits << 4) >> 12)
	//mask is bit 8
	mask = (currbits<<8)>>15 == 1
	//payload length (field 1) is bit 9-15
	data_len = uint64((currbits << 9) >> 9)

	if data_len == 126 {
		rcv_len, err = c.Read(buf)
		if err != nil || rcv_len != len(buf) {
			return retDG, errors.New(ERRMSG_NETWORK_READ + " occurrence 2")
		}
		data_len = binary.BigEndian.Uint64(buf)
	} else if data_len == 127 {
		buf = make([]byte, 8)
		rcv_len, err = c.Read(buf)
		if err != nil || rcv_len != len(buf) {
			return retDG, errors.New(ERRMSG_NETWORK_READ + " occurrence 3")
		}
		data_len = binary.BigEndian.Uint64(buf)
	}

	if mask {
		//(un)masking algorithm section 5.3
		//https://tools.ietf.org/html/rfc6455#section-5.3
		mask_key = make([]byte, 4)
		rcv_len, err = c.Read(mask_key)
		if err != nil || rcv_len != len(mask_key) {
			return retDG, errors.New(ERRMSG_NETWORK_READ + " occurrence 4")
		}
	}

	data = make([]byte, data_len)
	rcv_len, err = c.Read(data)
	if err != nil || rcv_len != len(data) {
		return retDG, errors.New(ERRMSG_NETWORK_READ + " occurrence 5")
	}

	if mask {
		//(un)masking algorithm section 5.3
		//https://tools.ietf.org/html/rfc6455#section-5.3
		for i := uint64(0); i < data_len; i++ {
			data[i] = data[i] ^ mask_key[i%4]
		}
	}

	retDG = Websocket_datagram{fin, opcode, data_len, data}
	return retDG, nil
}

func NewWebsocketDatagram(d []byte, contenttype int) (Websocket_datagram, error) {
	var retdg Websocket_datagram
	var datacopy []byte

	if contenttype != websocket_type_text && contenttype != websocket_type_binary {
		return retdg, errors.New(ERRMSG_INVALID_ENUM)
	}

	datacopy = make([]byte, len(d))
	copy(datacopy, d)
	retdg = Websocket_datagram{Fin: true, Opcode: byte(contenttype), data_len: uint64(len(d)), Data: datacopy}
	return retdg, nil
}

func (dg *Websocket_datagram) CreateDatagram() ([]byte, error) {
	var err error
	var bytebuf []byte
	var data []byte
	var currbyte uint8
	var currbyteindex uint64
	var mask_key uint32
	var mask_key_buf []byte

	if len(dg.Data) == 0 {
		return data, errors.New(ERRMSG_ZEROLENGTH)
	}

	//we add 2 extra bytes for the size, plus 4 bytes for the mask
	var sz uint64 = dg.data_len + 6

	//calculate frame size: assumption: b cannot be longer than an int64 can represent
	if sz > 125 {
		if sz < 2<<16 {
			//spec says so
			//https://tools.ietf.org/html/rfc6455#section-5.2
			sz += 2
		} else {
			//spec says so
			//https://tools.ietf.org/html/rfc6455#section-5.2
			sz += 8
		}
	}

	data = make([]byte, sz)

	//trusting that opcode is correct, will mask
	currbyte = uint8(2)<<6 + dg.Opcode
	data[0] = currbyte

	//encode size and mask
	//https://tools.ietf.org/html/rfc6455#section-5.2
	if dg.data_len < 126 {
		currbyte = uint8(dg.data_len)
		data[1] = currbyte + uint8(2)<<6
		currbyteindex = 2
	} else if dg.data_len < 2<<16 {
		currbyte = uint8(126)
		data[1] = currbyte + uint8(2)<<6
		bytebuf = make([]byte, 2)
		binary.BigEndian.PutUint16(bytebuf, uint16(dg.data_len))
		for i := 0; i < 2; i++ {
			data[i+2] = bytebuf[i]
		}
		currbyteindex = 4
	} else {
		currbyte = uint8(127) + uint8(2)<<6
		data[1] = currbyte
		bytebuf = make([]byte, 8)
		binary.BigEndian.PutUint64(bytebuf, uint64(dg.data_len))
		for i := 0; i < 8; i++ {
			data[i+2] = bytebuf[i]
		}
		currbyteindex = 10
	}

	//make and write mask key
	mask_key = rand.Uint32()
	mask_key_buf = make([]byte, 4)
	binary.BigEndian.PutUint32(mask_key_buf, mask_key)
	for i := uint64(0); i < 4; i++ {
		data[currbyteindex+i] = mask_key_buf[i]
	}
	currbyteindex += 4

	for i := uint64(0); i < dg.data_len; i++ {
		data[currbyteindex+i] = dg.Data[i] ^ mask_key_buf[i%4]
	}
	log.Printf("len data = %ld\tdataframe %v\n", dg.data_len, data)
	return data, err
}

//SendStruct is used to send **JSON-MARSHALLED structs over a network connection
//i should be a pointer
func SendStruct(i interface{}, c net.Conn) error {
	const length_of_int = 4
	var b []byte
	var err error
	var length uint32
	var recvlength int
	var content_lenbuf []byte
	var client_lenbuf []byte
	var errbuf []byte

	if i == nil {
		return errors.New(ERRMSG_NILPTR)
	}
	b, err = json.Marshal(i)

	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return err
	}
	content_lenbuf = make([]byte, length_of_int)
	client_lenbuf = make([]byte, length_of_int)
	errbuf = make([]byte, 1)

	length = uint32(len(b))
	binary.BigEndian.PutUint32(content_lenbuf, length)

	//COMPLIMENTARY: Send length of JSON object to send
	recvlength, err = c.Write(content_lenbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_WRITE)
	if recvlength != length_of_int {
		return errors.New(ERRMSG_NETWORK_WRITE)
	}
	//COMPLIMENTARY: Receive same length as confirmation of receipt
	c.Read(client_lenbuf)
	if bytes.Equal(content_lenbuf, client_lenbuf) {
		//COMPLIMENTARY: Send json struct
		_, err = c.Write(b)
		return err
	} else {
		//COMPLIMENTARY: Send error message
		errbuf[0] = NETCODE_ERR
		c.Write(errbuf)
		return errors.New(ERRMSG_NETWORK_WRITE)
	}
}

func SendStruct_JSClient(i interface{}, c net.Conn) error {
	var b []byte
	var err error
	var sentdg Websocket_datagram

	b, err = json.Marshal(i)
	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return err
	}
	sentdg, err = NewWebsocketDatagram(b, websocket_type_binary)
	Errhandle_Log(err, ERRMSG_CREATE_DATAGRAM_STRUCT)
	if err != nil {
		return err
	}

	b, err = sentdg.CreateDatagram()
	Errhandle_Log(err, ERRMSG_CREATE_DATAGRAM)
	if err != nil {
		return err
	}

	n, err := c.Write(b)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if n != len(b) || err != nil {
		return err
	}
	return nil
}

//complement to SendStruct
//PASS A POINTER TO RECEIVE THE STRUCT
func ReceiveStruct(i interface{}, c net.Conn) error {

	const length_of_int = 4
	var intbuf []byte = make([]byte, length_of_int)
	var contentlength int
	var contentbuf []byte
	var recvlength int
	var err error

	//COMPLIMENTARY: Receive length of JSON object to make buffer
	_, err = c.Read(intbuf)
	if err != nil {
		Errhandle_Log(err, ERRMSG_NETWORK_READ)
	}
	contentlength = int(binary.BigEndian.Uint32(intbuf))
	log.Printf("CONTENT LENGTH=%d\n", contentlength)
	contentbuf = make([]byte, contentlength)
	//COMPLIMENTARY: Send received length as confirmation of receipt
	_, err = c.Write(intbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_WRITE)
	//COMPLIMENTARY: Receive either JSON struct or error message
	recvlength, err = c.Read(contentbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_READ)
	//error message or bad transmission
	if recvlength != contentlength {
		log.Printf("Expected read length: %d\ttrue read length: %d, received message: %s",
			len(contentbuf), recvlength, contentbuf)
		return errors.New(ERRMSG_NETWORK_READ)
	}
	//should work unless wrong struct passed in
	err = json.Unmarshal(contentbuf, i)
	return err
}

func ReceiveStruct_JSClient(i interface{}, c net.Conn) error {
	var datagram Websocket_datagram
	var err error

	datagram, err = datagram.ReadDatagram(c)

	if err != nil {
		Errhandle_Log(err, ERRMSG_NETWORK_DATAGRAM)
		return err
	}
	log.Printf("%v\n", datagram)
	err = json.Unmarshal(datagram.Data, i)
	if err != nil {
		Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
		return err
	}
	return nil
}

func NewNetmessage(message string) Netmessage {
	return Netmessage{message}
}
