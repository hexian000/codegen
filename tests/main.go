package main

import "fmt"

type B struct {
	i int
	f float64
	c complex128
}
type C rune

//go:generate go run ../serialize.go -type=A,B,C
type A struct {
	B
	a   [3]B
	s   string
	ss  []string
	sss [][]string
	ms  map[string]string
	mms map[string]map[string]string
	si  struct {
		uint64
	}
	C
}

func main() {
	fmt.Println(A{})
}
