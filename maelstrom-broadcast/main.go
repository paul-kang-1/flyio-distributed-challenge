package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type msgBody map[string]any

var (
	db         map[int]msgBody
	neighbors  []string
	retries    map[string][]map[string]any
	dbMutex    sync.RWMutex
	retryMutex sync.RWMutex
)

func read_db(key int) (map[string]any, bool) {
	dbMutex.RLock()
	val, ok := db[key]
	dbMutex.RUnlock()
	return val, ok
}

func write_db(key int, value *msgBody) {
	dbMutex.Lock()
	db[key] = *value
	dbMutex.Unlock()
}

func retry(n *maelstrom.Node) {
	for recipient, msgs := range retries {
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
		_, ok := read_db(message)
		if !ok {
			dbMutex.Lock()
			db[message] = nil
			dbMutex.Unlock()
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
		dbMutex.RLock()
		res := make([]int, len(db))
		i := 0
		for k := range db {
			res[i] = k
			i++
		}
		dbMutex.RUnlock()
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
	db = make(map[int]msgBody)
	dbMutex = sync.RWMutex{}
	retryMutex = sync.RWMutex{}

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
