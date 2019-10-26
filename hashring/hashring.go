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

func (hr *HashRing) Init() *HashRing {
	hr.nodeValue2ID = map[string]int32{}
	hr.nodeID2Value = map[int32]string{}
	return hr
}

func (hr *HashRing) AddNode(nodeValue string, nodeWeight int32) bool {
	if _, ok := hr.nodeValue2ID[nodeValue]; ok {
		return false
	}

	nodeID := hr.getNextNodeID()
	hr.nodeValue2ID[nodeValue] = nodeID
	hr.nodeID2Value[nodeID] = nodeValue

	for i := int32(0); i < nodeWeight; i++ {
		h := fnv.New32a()
		fmt.Fprintf(h, "%d-%s", i, nodeValue)

		hr.nodes = append(hr.nodes, hashRingNode{
			Sum: h.Sum32(),
			ID:  nodeID,
		})
	}

	sort.Slice(hr.nodes, func(i, j int) bool {
		return hr.nodes[i].Sum >= hr.nodes[j].Sum
	})

	return true
}

func (hr *HashRing) RemoveNode(nodeValue string) bool {
	nodeID, ok := hr.nodeValue2ID[nodeValue]

	if !ok {
		return false
	}

	delete(hr.nodeValue2ID, nodeValue)
	delete(hr.nodeID2Value, nodeID)
	i := 0

	for _, node := range hr.nodes {
		if node.ID == nodeID {
			continue
		}

		hr.nodes[i] = node
		i++
	}

	hr.nodes = hr.nodes[:i]
	return true
}

func (hr *HashRing) FindNode(nodeKey string) (string, bool) {
	n := len(hr.nodes)

	if n == 0 {
		return "", false
	}

	h := fnv.New32a()
	h.Write([]byte(nodeKey))
	sum := h.Sum32()

	i := sort.Search(n, func(i int) bool {
		return hr.nodes[i].Sum <= sum
	})

	if i == n {
		i = 0
	}

	nodeValue := hr.nodeID2Value[hr.nodes[i].ID]
	return nodeValue, true
}

func (hr *HashRing) getNextNodeID() int32 {
	nodeID := hr.nextNodeID

	for n := math.MaxInt32; n >= 1; n-- {
		nextNodeID := int32((uint32(nodeID) + 1) & math.MaxInt32)

		if _, ok := hr.nodeID2Value[nodeID]; !ok {
			hr.nextNodeID = nextNodeID
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
