package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const USE_TOPOLOGY_MSG = false
const NUM_CHILD = 5

type (
	mapStruct[K comparable, V any] struct {
		sync.RWMutex
		m map[K]V
	}
	msgBody map[string]any
)

var (
	neighbors []string
	db        mapStruct[int, msgBody]
)

func (m *mapStruct[K, V]) Get(key K) (value V, ok bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok = m.m[key]
	return
}

func (m *mapStruct[K, V]) Put(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *mapStruct[K, V]) Keys() *[]K {
	m.RLock()
	defer m.RUnlock()
	res := make([]K, len(m.m))
	i := 0
	for k := range m.m {
		res[i] = k
		i++
	}
	return &res
}

func (m *mapStruct[K, V]) Length() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func handle_broadcast(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		var ackBody msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		if err := json.Unmarshal(msg.Body, &ackBody); err != nil {
			return err
		}
		message := int(ackBody["message"].(float64))
		delete(ackBody, "message")
		ackBody["type"] = "broadcast_ok"
		if err := n.Reply(msg, ackBody); err != nil {
			return err
		}
		_, ok := db.Get(message)
		if ok {
			return nil // return if message is already handled
		}
		db.Put(message, nil)
		waiting := sync.Map{}
		for _, neighbor := range neighbors {
			if neighbor == msg.Src {
				continue
			}
			waiting.Store(neighbor, false)
		}
		pending := true
		for pending {
			pending = false
			waiting.Range(func(neighbor, value any) bool {
				if v, _ := waiting.Load(neighbor); v.(bool) {
					return true
				}
				pending = true
				n.RPC(neighbor.(string), body, func(msg maelstrom.Message) error {
					waiting.Store(neighbor, true)
					return nil
				})
				return true
			})
			if pending {
				time.Sleep(time.Millisecond * 500)
			}
		}
		return nil
	}
}

func handle_read(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		res := *db.Keys()
		body["messages"] = res
		body["type"] = "read_ok"
		return n.Reply(msg, body)
	}
}

func get_topology(body *msgBody) (map[string][]string, error) {
	field, ok := (*body)["topology"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Invalid JSON body format: 'topology'")
	}
	topology := make(map[string][]string)
	for node, neighbors := range field {
		if neighborsJSON, ok := neighbors.([]any); ok {
			for _, neighbor := range neighborsJSON {
				if neighborStr, ok := neighbor.(string); ok {
					topology[node] = append(topology[node], neighborStr)
				}
			}
		}
	}
	return topology, nil
}

func handle_topology(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "topology_ok"
		if USE_TOPOLOGY_MSG {
			top, err := get_topology(&body)
			if err != nil {
				return err
			}
			neighbors = top[n.ID()]
		} else {
			spanning_tree_neighbor(n, NUM_CHILD)
		}
		delete(body, "topology")
		return n.Reply(msg, body)
	}
}

func spanning_tree_neighbor(n *maelstrom.Node, num_child int) {
	topology := make(map[string][]string)
	for _, node := range n.NodeIDs() {
		topology[node] = make([]string, 0)
	}
	queue := []int{0}
	idx := 1
	currNodeIdx := 0
	for len(queue) > 0 {
		currNodeIdx = queue[0]
		currNodeID := n.NodeIDs()[currNodeIdx]
		queue = queue[1:]
		for i := 0; i < num_child && idx < len(n.NodeIDs()); i++ {
			queue = append(queue, idx)
			neighborNodeID := n.NodeIDs()[idx]
			topology[currNodeID] = append(topology[currNodeID], neighborNodeID)
			topology[neighborNodeID] = append(topology[neighborNodeID], currNodeID)
			idx++
		}
	}
	neighbors = topology[n.ID()]
}

func main() {
	n := maelstrom.NewNode()
	db.RWMutex = sync.RWMutex{}
	db.m = make(map[int]msgBody)

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
