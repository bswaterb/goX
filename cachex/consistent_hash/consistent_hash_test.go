package consistent_hash

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	h := NewConsistentHashMap(nil)

	h.AddNodes(Node{
		Id:              "node1",
		VirtualNodeNums: 1,
	})

	h.AddNodes(Node{
		Id:              "node2",
		VirtualNodeNums: 2,
	})

	for i := 0; i < 100; i++ {
		k := strconv.Itoa(i)
		v := fmt.Sprintf("这是 %s 的值: %d", k, rand.Int())
		h.Set(k, v)
	}

	for i := 0; i < 100; i++ {
		k := strconv.Itoa(i)
		v, err := h.Get(k)
		assert.NoError(t, err)
		fmt.Println(v)
	}

}
