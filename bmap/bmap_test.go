package bmap

import (
	"fmt"
	"testing"
)

func TestBMap(t *testing.T) {
	m := NewBMap(WithHashFunc(func(key string, len int) int {
		return 44
	}))

	fmt.Println(m.Get("1"))
	m.Store("1", "123")
	fmt.Println(m.Get("1"))
	m.Store("2", "456")
	fmt.Println(m.Get("2"))
	fmt.Println(m.Get("1"))

}
