package main

import (
	"errors"
	"fmt"
	"net"
)

func selectServerIP() error {
	var userinput string
	var userinputsz int
	var cutindex int
	var prebyteslice []byte
	var err error
	//significant code adopted from:
	//https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
	ifaces, err := net.Interfaces()
	if err != nil {
		Errhandle_Exit(err, ERRMSG_FETCH_IP_TABLE)
	}
	for c, i := range ifaces {
		fmt.Printf("iteration %d\n", c)
		addrs, err := i.Addrs()
		if err != nil {
			Errhandle_Exit(err, ERRMSG_FETCH_IP_TABLE)
		}
		for i, addr := range addrs {
			fmt.Printf("Server IP Address Option %d: %v\n Use? (y/n) ", i, addr.String())
			_, err = fmt.Scanf("%s\n", &userinput)
			Errhandle_Exit(err, ERRMSG_INPUT_RETRIEVAL)
			fmt.Printf("scanned! User input = %s\n", userinput)
			if userinput != "y" && userinput != "Y" {
				if c == len(ifaces)-1 && i == len(addrs)-1 {
					return errors.New(ERRMSG_FETCH_IP_TABLE_NONE_SELECTED)
				}
				continue
			}
			userinput = addr.String()
			userinputsz = len(userinput)
			for i := 0; i < userinputsz; i++ {
				if userinput[i] == '/' {
					cutindex = i
					break
				}
			}
			prebyteslice = make([]byte, cutindex+len(PORT))
			for i := 0; i < cutindex; i++ {
				prebyteslice[i] = userinput[i]
			}
			for i := cutindex; i < len(prebyteslice); i++ {
				prebyteslice[i] = PORT[i-cutindex]
			}
			fmt.Printf("pbs: %v\n, uis: %d", prebyteslice, userinputsz)
			IP_SERVER = string(prebyteslice)
			return nil
		}
	}
	return nil
}
