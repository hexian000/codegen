package main

import (
	"encoding/binary"
	"testing"
)

var a = A{}

func BenchmarkA_Serialize(b *testing.B) {
	l := a.SerializeLen()
	data := make([]byte, l, l)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Serialize(data, binary.BigEndian)
	}
}

func BenchmarkA_Deserialize(b *testing.B) {
	l := a.SerializeLen()
	data := make([]byte, l, l)
	a.Serialize(data, binary.BigEndian)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Deserialize(data, binary.BigEndian)
	}
}
