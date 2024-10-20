package bmap

import "hash/fnv"

const (
	defaultSlot = 100
)

type BMap struct {
	hashFunc func(key string, len int) int
	// slot number
	len int
	// k-v
	data []*Node
}

type Node struct {
	k string
	v any
	// Separate chaining
	next *Node
}

type Opt func(bMap *BMap)

func defaultHashFunc(key string, len int) int {
	// return hash(key) % m.len
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % len
}

func WithHashFunc(hashFunc func(key string, len int) int) Opt {
	return func(bMap *BMap) {
		bMap.hashFunc = hashFunc
	}
}

func NewBMap(opts ...Opt) *BMap {
	m := &BMap{}
	m.hashFunc = defaultHashFunc
	m.len = defaultSlot
	m.data = make([]*Node, m.len)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *BMap) Get(key string) (any, bool) {
	hash := m.hashFunc(key, m.len)
	node := m.data[hash]
	for node != nil {
		if node.k == key {
			return node.v, true
		}
		node = node.next
	}
	return nil, false
}

func (m *BMap) Store(key string, value any) {
	hash := m.hashFunc(key, m.len)

	if m.data[hash] == nil {
		m.data[hash] = &Node{k: key, v: value}
		return
	}

	node := m.data[hash]
	for node.next != nil {
		if node.k == key {
			node.v = value
			return
		}
		node = node.next
	}

	node.next = &Node{k: key, v: value}
}
