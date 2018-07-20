package main

import "testing"

var a = A{}

func BenchmarkA_Serialize(b *testing.B) {
	data := make([]byte, 0, 65536)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Serialize(data)
	}
}

func BenchmarkA_Deserialize(b *testing.B) {
	data := a.Serialize(make([]byte, 0, 65536))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Deserialize(data)
	}
}
