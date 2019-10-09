package hashring

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"unsafe"
)

type HashRing struct {
	nodeValue2ID map[string]int32
	nodeID2Value map[int32]string
	nextNodeID   int32
	nodes        []hashRingNode
}

func (self *HashRing) Init() *HashRing {
	self.nodeValue2ID = map[string]int32{}
	self.nodeID2Value = map[int32]string{}
	return self
}

func (self *HashRing) AddNode(nodeValue string, nodeWeight int32) bool {
	if _, ok := self.nodeValue2ID[nodeValue]; ok {
		return false
	}

	nodeID := self.getNextNodeID()
	self.nodeValue2ID[nodeValue] = nodeID
	self.nodeID2Value[nodeID] = nodeValue

	for i := int32(0); i < nodeWeight; i++ {
		h := fnv.New32a()
		fmt.Fprintf(h, "%d-%s", i, nodeValue)

		self.nodes = append(self.nodes, hashRingNode{
			Sum: h.Sum32(),
			ID:  nodeID,
		})
	}

	sort.Slice(self.nodes, func(i, j int) bool {
		return self.nodes[i].Sum >= self.nodes[j].Sum
	})

	return true
}

func (self *HashRing) RemoveNode(nodeValue string) bool {
	nodeID, ok := self.nodeValue2ID[nodeValue]

	if !ok {
		return false
	}

	delete(self.nodeValue2ID, nodeValue)
	delete(self.nodeID2Value, nodeID)
	i := 0

	for _, node := range self.nodes {
		if node.ID == nodeID {
			continue
		}

		self.nodes[i] = node
		i++
	}

	self.nodes = self.nodes[:i]
	return true
}

func (self *HashRing) FindNode(nodeKey string) (string, bool) {
	n := len(self.nodes)

	if n == 0 {
		return "", false
	}

	h := fnv.New32a()
	h.Write([]byte(nodeKey))
	sum := h.Sum32()

	i := sort.Search(n, func(i int) bool {
		return self.nodes[i].Sum <= sum
	})

	if i == n {
		i = 0
	}

	nodeValue := self.nodeID2Value[self.nodes[i].ID]
	return nodeValue, true
}

func (self *HashRing) getNextNodeID() int32 {
	nodeID := self.nextNodeID

	for n := math.MaxInt32; n >= 1; n-- {
		nextNodeID := int32((uint32(nodeID) + 1) & math.MaxInt32)

		if _, ok := self.nodeID2Value[nodeID]; !ok {
			self.nextNodeID = nextNodeID
			return nodeID
		}

		nodeID = nextNodeID
	}

	panic(errors.New("hashring: too many nodes"))
}

type hashRingNode struct {
	Sum uint32
	ID  int32
}

var _ [unsafe.Sizeof(hashRingNode{})]struct{} = [unsafe.Sizeof(int64(0))]struct{}{}
