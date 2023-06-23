package main

import (
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

var kv *maelstrom.KV

func handleAdd() maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
}

func handleRead() maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
}

func main() {
	n := maelstrom.NewNode()
	kv = maelstrom.NewSeqKV(n)

	n.Handle("add", handleAdd())
	n.Handle("read", handleRead())

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
