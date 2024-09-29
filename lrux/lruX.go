package lrux

import (
	"container/list"
	"log"
	"sync"
	"time"
)

var upgradeTimeThreshold = 2 * time.Second

type cacheItem struct {
	key        string
	value      interface{}
	insertTime time.Time
	inOld      bool
}

type LRUCache struct {
	capacityOld   int
	capacityYoung int
	listOld       *list.List
	listYoung     *list.List
	items         map[string]*list.Element
	mutex         sync.Mutex
}

func NewLRUCache(capacityOld, capacityYoung int) *LRUCache {
	return &LRUCache{
		capacityOld:   capacityOld,
		capacityYoung: capacityYoung,
		listOld:       list.New(),
		listYoung:     list.New(),
		items:         make(map[string]*list.Element),
	}
}

// Get 从 LRU 链表中通过 key 获取 value
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, found := c.items[key]; found {
		item := elem.Value.(*cacheItem)
		// 如果该元素已经在 young 区存在了 upgradeTimeThreshold 的时长，那么再次被访问时将其转移到 old 区中
		if !item.inOld && time.Since(item.insertTime) > upgradeTimeThreshold {
			c.promoteToOld(key)
		} else if !item.inOld {
			c.listYoung.MoveToFront(elem)
		} else {
			c.listOld.MoveToFront(elem)
		}
		return item.value, true
	}
	return nil, false
}

// Set 在 LRU 链表中添加或更新 k-v
func (c *LRUCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 若 key 已存在于 LRU 链表中，则更新其值
	if elem, found := c.items[key]; found {
		if !elem.Value.(*cacheItem).inOld {
			c.listYoung.MoveToFront(elem)
		} else {
			c.listOld.MoveToFront(elem)
		}
		elem.Value.(*cacheItem).value = value
		return
	}

	// 如果 young 区已满，淘汰最久没被访问的那个元素
	if c.listYoung.Len() >= c.capacityYoung {
		c.evict(c.listYoung, "young")
	}

	// 新来的元素默认被添加到 young 区
	item := &cacheItem{
		key:        key,
		value:      value,
		insertTime: time.Now(),
	}
	elem := c.listYoung.PushFront(item)
	c.items[key] = elem
}

// promoteToOld 将 k-v 从 young 链表转移到 old 链表
func (c *LRUCache) promoteToOld(key string) {
	log.Default().Println("有元素触发 young -> old 转移：", key)
	elem, found := c.items[key]
	if !found {
		return
	}

	item := elem.Value.(*cacheItem)
	// 从 young 区摘除当前节点
	c.listYoung.Remove(elem)

	if c.listOld.Len() >= c.capacityOld {
		c.evict(c.listOld, "old")
	}

	// 添加到 old 区中
	item.inOld = true
	elem = c.listOld.PushFront(item)
	c.items[key] = elem
}

// evict 淘汰指定链表中的最久没被访问过的元素
func (c *LRUCache) evict(list *list.List, listType string) {
	if elem := list.Back(); elem != nil {
		item := elem.Value.(*cacheItem)
		list.Remove(elem)
		delete(c.items, item.key)
		log.Default().Println("有元素触发了淘汰：", listType, item.key)
	}
}

func (c *LRUCache) eleInOld(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if v, exists := c.items[key]; exists {
		ele := v.Value.(*cacheItem)
		return ele.inOld
	}
	return false
}
