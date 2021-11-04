package main

/*
	Description:
	if a and b are of equal length, xor will return a slice where
		retslice[i] = a[i]^b[i] for i < len(a)

	if a and b are not of equal size, suppose len(a) < len(b).
		xor will let retslice[i] = a[i] ^ b[i] if i < len(a)
		xor will let retslice[i] = 0 if i > len(a)
*/
func xor(a, b []byte) (retslice []byte) {
	if a == nil && b == nil {
		return make([]byte, 1)
	}
	if len(a) == len(b) {
		retslice = make([]byte, len(a))
		for i := 0; i < len(a); i++ {
			retslice[i] = a[i] ^ b[i]
		}
	} else {
		if len(a) > len(b) {
			var i int
			retslice = make([]byte, len(a))
			for i = 0; i < len(b); i++ {

				retslice[i] = a[i] ^ b[i]
			}
			for ; i < len(a); i++ {

				retslice[i] = a[i]
			}
		} else {
			retslice = make([]byte, len(b))
			var i int
			for i = 0; i < len(a); i++ {

				retslice[i] = a[i] ^ b[i]
			}
			for ; i < len(b); i++ {

				retslice[i] = b[i]
			}
		}
	}
	return
}
