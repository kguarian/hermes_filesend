package main

import (
	"C"
)

//export MakeGoString
func MakeGoString(c *C.char) string {
	return string(*c)
}
