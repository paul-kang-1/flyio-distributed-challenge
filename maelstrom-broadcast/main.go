package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var db map[int]struct{}

func handle_broadcast(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		message := int(body["message"].(float64))
		db[message] = struct{}{}
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

func handle_topology(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "topology_ok"
		delete(body, "topology")
		return n.Reply(msg, body)
	}
}

func main() {
	n := maelstrom.NewNode()
	db = make(map[int]struct{})

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
