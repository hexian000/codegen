package main

import "unsafe"

var _ = unsafe.Sizeof(0)

func (i *C) Serialize(b []byte) []byte {
	b = append(b, byte(*i>>24), byte(*i>>16), byte(*i>>8), byte(*i))
	return b
}

func (i *C) Deserialize(b []byte) []byte {
	f := rune(b[3]) | rune(b[2])<<8 | rune(b[1])<<16 | rune(b[0])<<24
	b = b[4:]
	*i = C(f)
	return b
}
