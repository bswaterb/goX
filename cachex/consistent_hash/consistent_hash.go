package consistent_hash

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
)

var (
	ErrHashNotInit = errors.New("哈希环未初始化")
)

type Hash func(data []byte) uint32

type Node struct {
	Id              string
	VirtualNodeNums int
	kvStorage       *NodeStorage
}

type NodeStorage struct {
	db map[string]any
}

type Map struct {
	hashFunc Hash

	// sortedHashes nodes hash result in hash-ring
	sortedHashes []int

	// k: node_id v: node_obj
	// for example:
	// k: node1_v1_id v: node1
	// k: node1_v2_id v: node1
	// k: node2_v1_id v: node2
	nodesMap map[int]*Node
}

func NewConsistentHashMap(fn Hash) *Map {
	m := &Map{
		hashFunc:     fn,
		sortedHashes: make([]int, 0, 10),
		nodesMap:     make(map[int]*Node),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) AddNodes(nodes ...Node) {
	for i := range nodes {
		node := nodes[i]
		kvDb := &NodeStorage{
			db: make(map[string]any),
		}
		node.kvStorage = kvDb
		for i := 0; i < node.VirtualNodeNums; i++ {
			nodeVId := fmt.Sprintf("%s_v%d", node.Id, i)
			hashResult := m.hashFunc([]byte(nodeVId))
			m.nodesMap[int(hashResult)] = &node
			m.sortedHashes = append(m.sortedHashes, int(hashResult))
		}
	}
	sort.Ints(m.sortedHashes)
}

func (m *Map) Set(key string, value any) {
	if len(m.sortedHashes) == 0 {
		fmt.Println("哈希环中不包含数据节点，请先添加节点以初始化一致性哈希结构")
		return
	}
	keyHash := m.hashFunc([]byte(key))

	mappingIdx := sort.Search(len(m.sortedHashes), func(i int) bool {
		return m.sortedHashes[i] >= int(keyHash)
	})
	if mappingIdx == len(m.sortedHashes) {
		mappingIdx = 0
	}

	nodeHash := m.sortedHashes[mappingIdx]
	node := m.nodesMap[nodeHash]
	fmt.Println("当前 key 命中的节点是: ", node.Id)
	node.kvStorage.db[key] = value
}

func (m *Map) Get(key string) (any, error) {
	if len(m.sortedHashes) == 0 {
		return "", ErrHashNotInit
	}

	keyHash := m.hashFunc([]byte(key))

	mappingIdx := sort.Search(len(m.sortedHashes), func(i int) bool {
		return m.sortedHashes[i] >= int(keyHash)
	})

	if mappingIdx == len(m.sortedHashes) {
		mappingIdx = 0
	}

	nodeHash := m.sortedHashes[mappingIdx]
	node := m.nodesMap[nodeHash]
	fmt.Println("当前 key 命中的节点是: ", node.Id)

	return node.kvStorage.db[key], nil
}
