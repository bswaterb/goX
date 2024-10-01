package bttp

import "strings"

type treeNode struct {
	// 只有路径的最终节点才会填充此处的 pattern
	pattern string
	// 路径 '/' 分隔的当前子块
	subPath string
	// 子节点
	childNodes []*treeNode
	// 当前节点是否是精准匹配的节点（即此 subPath 不包含通配符/路径参数）
	exactMatch bool
}

func (n *treeNode) insert(pattern string, subPaths []string, idx int) {
	if len(subPaths) == idx {
		if len(subPaths) == 0 {
			n.exactMatch = true
		}
		n.pattern = pattern
		return
	}

	subPath := subPaths[idx]
	nextNode := n.matchChild(subPath)

	if nextNode == nil {
		nextNode = &treeNode{subPath: subPath, exactMatch: true}
		if len(subPath) != 0 {
			nextNode.exactMatch = subPath[0] != ':' && subPath[0] != '*'
		}
		n.childNodes = append(n.childNodes, nextNode)
	}

	nextNode.insert(pattern, subPaths, idx+1)
}

func (n *treeNode) search(subPaths []string, idx int) *treeNode {
	if len(subPaths) == idx || strings.HasPrefix(n.subPath, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	subPath := subPaths[idx]
	children := n.matchChildren(subPath)

	for _, child := range children {
		result := child.search(subPaths, idx+1)
		if result != nil {
			return result
		}
	}

	return nil
}

// 第一个匹配成功的节点，用于插入
func (n *treeNode) matchChild(subPath string) *treeNode {
	for _, child := range n.childNodes {
		if child.subPath == subPath {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *treeNode) matchChildren(subPath string) []*treeNode {
	nodes := make([]*treeNode, 0)
	for _, child := range n.childNodes {
		if child.subPath == subPath || !child.exactMatch {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
