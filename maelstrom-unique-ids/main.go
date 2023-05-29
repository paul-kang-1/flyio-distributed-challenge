package main

import (
	"encoding/json"
	"log"
	"strconv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var count = 0

func generate(nodeId string) string {
	count++
	return nodeId + strconv.Itoa(count)
}

func main() {
	n := maelstrom.NewNode()
	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string] any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// Generate ID here
		body["id"] = generate(n.ID());
		body["type"] = "generate_ok"
		return n.Reply(msg, body)
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

