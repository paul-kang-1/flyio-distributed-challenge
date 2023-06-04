package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type (
	mapStruct[K comparable, V any] struct {
		sync.RWMutex
		m map[K]V
	}
	msgBody                        map[string]any
)

var (
	neighbors []string

	db      mapStruct[int, msgBody]
	retries mapStruct[string, []msgBody]
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

func retry(n *maelstrom.Node) {
	for recipient, msgs := range retries.m {
		for _, msg := range msgs {
			n.Send(recipient, msg)
		}
	}
}

func handle_broadcast(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		message := int(body["message"].(float64))
		_, ok := db.Get(message)
		if !ok {
			db.Put(message, nil)
			for _, neighbor := range neighbors {
				if neighbor != msg.Src {
					n.Send(neighbor, body)
				}
			}
		}
		delete(body, "message")
		body["type"] = "broadcast_ok"
		return n.Reply(msg, body)
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
	field, ok := (*body)["topology"].(msgBody)
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
		top, err := get_topology(&body)
		if err != nil {
			return err
		}
		neighbors = top[n.ID()]
		delete(body, "topology")
		return n.Reply(msg, body)
	}
}

func main() {
	n := maelstrom.NewNode()
	db.RWMutex = sync.RWMutex{}
	db.m = make(map[int]msgBody)
	retries.RWMutex = sync.RWMutex{}
	retries.m = make(map[string][]msgBody)

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
