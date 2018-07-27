// Code generated by serialize.
// source: main.go
// DO NOT EDIT!

package main

import "unsafe"

var _ = unsafe.Sizeof(0)

func (i *A) Serialize(b []byte) []byte {
	b = i.B.Serialize(b)
	for _, element := range i.a {
		b = element.Serialize(b)
	}
	{
		l := len(i.s)
		b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
		b = append(b, []byte(i.s)...)
	}
	{
		l := len(i.ss)
		b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
		for _, element := range i.ss {
			{
				l := len(element)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				b = append(b, []byte(element)...)
			}
		}
	}
	{
		l := len(i.sss)
		b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
		for _, element := range i.sss {
			{
				l := len(element)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				for _, element := range element {
					{
						l := len(element)
						b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
						b = append(b, []byte(element)...)
					}
				}
			}
		}
	}
	{
		l := len(i.ms)
		b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
		for key, value := range i.ms {
			{
				l := len(key)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				b = append(b, []byte(key)...)
			}
			{
				l := len(value)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				b = append(b, []byte(value)...)
			}
		}
	}
	{
		l := len(i.mms)
		b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
		for key, value := range i.mms {
			{
				l := len(key)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				b = append(b, []byte(key)...)
			}
			{
				l := len(value)
				b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
				for key, value := range value {
					{
						l := len(key)
						b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
						b = append(b, []byte(key)...)
					}
					{
						l := len(value)
						b = append(b, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
						b = append(b, []byte(value)...)
					}
				}
			}
		}
	}
	b = append(b, byte(i.si.uint64>>56), byte(i.si.uint64>>48), byte(i.si.uint64>>40), byte(i.si.uint64>>32), byte(i.si.uint64>>24), byte(i.si.uint64>>16), byte(i.si.uint64>>8), byte(i.si.uint64))
	b = i.C.Serialize(b)
	return b
}

func (i *A) Deserialize(b []byte) []byte {
	b = i.B.Deserialize(b)
	for k := uint(0); k < 3; k++ {
		b = i.a[k].Deserialize(b)
	}
	{
		l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
		b = b[4:]
		i.s = string(b[:l])
		b = b[l:]
	}
	{
		l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
		b = b[4:]
		i.ss = make([]string, l, l)
		b = b[l:]
		for k := uint(0); k < l; k++ {
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				i.ss[k] = string(b[:l])
				b = b[l:]
			}
		}
	}
	{
		l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
		b = b[4:]
		i.sss = make([][]string, l, l)
		b = b[l:]
		for k := uint(0); k < l; k++ {
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				i.sss[k] = make([]string, l, l)
				b = b[l:]
				for k := uint(0); k < l; k++ {
					{
						l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
						b = b[4:]
						i.sss[k][k] = string(b[:l])
						b = b[l:]
					}
				}
			}
		}
	}
	{
		l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
		b = b[4:]
		m := make(map[string]string)
		for k := uint(0); k < l; k++ {
			var key string
			var value string
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				key = string(b[:l])
				b = b[l:]
			}
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				value = string(b[:l])
				b = b[l:]
			}
			m[key] = value
		}
		i.ms = m
	}
	{
		l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
		b = b[4:]
		m := make(map[string]map[string]string)
		for k := uint(0); k < l; k++ {
			var key string
			var value map[string]string
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				key = string(b[:l])
				b = b[l:]
			}
			{
				l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
				b = b[4:]
				m := make(map[string]string)
				for k := uint(0); k < l; k++ {
					var key string
					var value string
					{
						l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
						b = b[4:]
						key = string(b[:l])
						b = b[l:]
					}
					{
						l := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
						b = b[4:]
						value = string(b[:l])
						b = b[l:]
					}
					m[key] = value
				}
				value = m
			}
			m[key] = value
		}
		i.mms = m
	}
	i.si.uint64 = uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 | uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	b = b[8:]
	b = i.C.Deserialize(b)
	return b
}
