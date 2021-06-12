package main

//USING BIG-ENDIAN
//REF: 	https://www.ibm.com/docs/en/error?originalUrl=SSB27U_6.4.0/com.ibm.zvm.v640.kiml0/asonetw.htm#:~:text=The%20network%20byte%20order%20is,confusion%20because%20of%20byte%20ordering.

type Netmessage struct {
	Message string
}

func NewNetmessage(message string) Netmessage {
	return Netmessage{message}
}
