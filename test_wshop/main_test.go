package main

import (
	"encoding/binary"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
)

func TestXorAlgorithm(t *testing.T) {
	var f *fuzz.Fuzzer
	var time_string_1, time_string_2 string
	var time_mark, time_mark_2 int64
	var time_slice_1, time_slice_2 []byte
	var fz, hash, input []byte

	f = fuzz.New()
	fz = make([]byte, 10)
	time_slice_1 = make([]byte, 8)
	time_mark = time.Now().Unix()
	time_string_1 = time.Unix(0, time_mark).Format(time.ANSIC)
	for testincrement := 0; testincrement < 100000; testincrement++ {
		f.Fuzz(&fz)
		binary.LittleEndian.PutUint64(time_slice_1, uint64(time_mark))
		hash = xor(time_slice_1, input)

		//so this works...

		//but not when we convert these to js strings? Doubt it

		time_slice_2 = xor(hash, input)[0:8]
		time_mark_2 = int64(binary.LittleEndian.Uint64(time_slice_2))
		time_string_2 = time.Unix(0, time_mark_2).Format(time.ANSIC)

		if time_string_1 != time_string_2 {
			t.Fail()
		}
	}
}
