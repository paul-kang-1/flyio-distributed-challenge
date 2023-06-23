package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var kv *maelstrom.KV

const counterKey = "key"

type msgBody map[string]any

func handleAdd(n *maelstrom.Node, ctx context.Context) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		var err error
		if err = json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		delta := int(body["delta"].(float64))

		var value int
		var ok bool
		var rpcErr *maelstrom.RPCError
		accepted := false
		for !accepted {
			value, err = read(n, ctx)
			if rpcErr, ok = err.(*maelstrom.RPCError); ok && rpcErr.Code == maelstrom.KeyDoesNotExist {
				value = 0
			}
			if err = kv.CompareAndSwap(ctx, counterKey, value, value+delta, true); err != nil {
				if rpcErr, ok = err.(*maelstrom.RPCError); ok && rpcErr.Code == maelstrom.PreconditionFailed {
					continue
				}
				return err
			}
			accepted = true
		}
		body["type"] = "add_ok"
		delete(body, "delta")
		return n.Reply(msg, &body)
	}
}

func handleRead(n *maelstrom.Node, ctx context.Context) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		var err error
		if err = json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		var value int
		value, err = read(n, ctx)
		if err != nil {
			return err
		}
		body["value"] = value
		body["type"] = "read_ok"
		return n.Reply(msg, body)
	}
}

func generateRandomKey(length int) string {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Fatalf("random key generation error: %+v", err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}

func read(n *maelstrom.Node, ctx context.Context) (int, error) {
	randomKey := generateRandomKey(24)
	var err error
	if err = kv.Write(ctx, randomKey, "dummy"); err != nil {
		return -1, err
	}
	var value int
	value, err = kv.ReadInt(ctx, counterKey)
	if err != nil {
		return -1, err
	}
	return value, nil
}

func main() {
	n := maelstrom.NewNode()
	kv = maelstrom.NewSeqKV(n)
	ctx := context.Background()

	n.Handle("add", handleAdd(n, ctx))
	n.Handle("read", handleRead(n, ctx))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
