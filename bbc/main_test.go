package main

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestXorAlgorithm(t *testing.T) {
	var time_string_1, time_string_2 string
	var time_mark, time_mark_2 int64
	var time_slice_1, time_slice_2 []byte
	var input []byte
	var hash []byte

	time_mark = time.Now().Unix()
	time_string_1 = time.Unix(0, time_mark).Format(time.ANSIC)
	input = []byte("No \"Who cares,\", no vacant stares, no time for me.")
	binary.LittleEndian.PutUint64(time_slice_1, uint64(time_mark))
	hash = xor(time_slice_1, input)
	time_slice_2 = xor(hash, input)[0:8]
	time_mark_2 = int64(binary.LittleEndian.Uint64(time_slice_2))
	time_string_2 = time.Unix(0, time_mark_2).Format(time.ANSIC)

	if time_string_1 != time_string_2 {
		t.Fail()
	}
}
