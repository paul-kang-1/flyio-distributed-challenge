package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func handle_broadcast(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		return n.Reply(msg, body)
	}
}

func handle_read(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		return n.Reply(msg, body)
	}
}

func handle_topology(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		return n.Reply(msg, body)
	}
}

func main() {
	n := maelstrom.NewNode()

	// Register handlers
	n.Handle("broadcast", handle_broadcast(n))
	n.Handle("read", handle_read(n))
	n.Handle("topology", handle_topology(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
