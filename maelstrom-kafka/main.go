package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

/*
 *  Request Handlers
 */

 func handleSend(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
 }

 func handlePoll(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
 }

 func handleCommitOffsets(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
 }

 func handleListCommitOffsets(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		return nil
	}
 }

func main() {
	n := maelstrom.NewNode()

	// Register handlers
	n.Handle("send", handleSend((n)))
	n.Handle("poll", handlePoll((n)))
	n.Handle("commit_offsets", handleCommitOffsets((n)))
	n.Handle("list_commit_offsets", handleListCommitOffsets((n)))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
