package util

import "fmt"

func BytesToU64(bytes []byte) (v uint64, e error) {
	if len(bytes) != 8 {
		e = fmt.Errorf("invalid input: %s", bytes)
		return
	}

	v = uint64(bytes[0])
	v <<= 8
	v += uint64(bytes[1])
	v <<= 8
	v += uint64(bytes[2])
	v <<= 8
	v += uint64(bytes[3])
	v <<= 8
	v += uint64(bytes[4])
	v <<= 8
	v += uint64(bytes[5])
	v <<= 8
	v += uint64(bytes[6])
	v <<= 8
	v += uint64(bytes[7])

	return
}

func BytesToI64(bytes []byte) (v int64, e error) {
	u, e := BytesToU64(bytes)
	v = int64(u)
	return
}

