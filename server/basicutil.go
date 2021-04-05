package main

func Trim(b []byte) []byte {
	var nonnullcount int = 0
	var retbuf []byte
	for _, c := range b {
		if c != 0 {
			nonnullcount += 1
			continue
		}
		break
	}
	retbuf = make([]byte, nonnullcount)
	for i := 0; i < nonnullcount; i++ {
		retbuf[i] = b[i]
	}
	return retbuf
}
