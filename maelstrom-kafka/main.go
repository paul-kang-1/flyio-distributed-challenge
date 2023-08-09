package main

import (
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"github.com/paul-kang-1/flyio-distributed-challenge/utils"
)

/*
 *  Data Structures
 */
 type Entry struct {
	offset int
	message any
 }

 var (
	 maxCommitted int
	 db utils.MapStruct[string, Entry]
 )

/*
 *  Request Handlers
 */

 func handleSend(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		key := body["key"].(string)
		value := body["msg"]
		var reply map[string]any
		return n.Reply(msg, reply)
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
	db = utils.MapStruct[string, Entry]{
		RWMutex: sync.RWMutex{},
		M: make(map[string]Entry),
	}
	// Register handlers
	n.Handle("send", handleSend((n)))
	n.Handle("poll", handlePoll((n)))
	n.Handle("commit_offsets", handleCommitOffsets((n)))
	n.Handle("list_commit_offsets", handleListCommitOffsets((n)))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
