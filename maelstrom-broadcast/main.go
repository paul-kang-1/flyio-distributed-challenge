package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var db map[int]struct{}
var topology map[string][]string

func handle_broadcast(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		message := int(body["message"].(float64))
		if _, ok := db[message]; !ok {
			db[message] = struct{}{}
			for _, neighbor := range topology[n.ID()] {
			// NOTE: Use below to ignore provided topology, assuming full connection
			// for _, neighbor := range n.NodeIDs() {
				n.Send(neighbor, body)
			}
		}
		delete(body, "message")
		body["type"] = "broadcast_ok"
		return n.Reply(msg, body)
	}
}

func handle_read(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		res := make([]int, len(db))
		i := 0
		for k := range db {
			res[i] = k
			i++
		}
		body["messages"] = res
		body["type"] = "read_ok"
		return n.Reply(msg, body)
	}
}

func get_topology(body *map[string]any) (map[string][]string, error) {
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
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "topology_ok"
		top, err := get_topology(&body)
		if err != nil {
			return err
		}
		topology = top
		delete(body, "topology")
		return n.Reply(msg, body)
	}
}

func main() {
	n := maelstrom.NewNode()
	db = make(map[int]struct{})
	topology = make(map[string][]string)

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
