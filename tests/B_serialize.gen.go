package main

import "unsafe"

var _ = unsafe.Sizeof(0)

func (i *B) Serialize(b []byte) []byte {
	b = append(b, byte(i.i>>56), byte(i.i>>48), byte(i.i>>40), byte(i.i>>32), byte(i.i>>24), byte(i.i>>16), byte(i.i>>8), byte(i.i))
	b = append(b, (*(*[unsafe.Sizeof(i.f)]byte)(unsafe.Pointer(&i.f)))[:]...)
	b = append(b, (*(*[unsafe.Sizeof(i.c)]byte)(unsafe.Pointer(&i.c)))[:]...)
	return b
}

func (i *B) Deserialize(b []byte) []byte {
	i.i = int(b[7]) | int(b[6])<<8 | int(b[5])<<16 | int(b[4])<<24 | int(b[3])<<32 | int(b[2])<<40 | int(b[1])<<48 | int(b[0])<<56
	b = b[8:]
	i.f = *(*float64)(unsafe.Pointer(&b[0]))
	b = b[unsafe.Sizeof(i.f):]
	i.c = *(*complex128)(unsafe.Pointer(&b[0]))
	b = b[unsafe.Sizeof(i.c):]
	return b
}
