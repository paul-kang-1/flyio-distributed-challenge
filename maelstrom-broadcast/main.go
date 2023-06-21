package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"github.com/paul-kang-1/flyio-distributed-challenge/utils"
)

const UseTopologyMsg = false
const UseSync = true
const NumChild = 5
const SyncMs = 250
const RetryMs = 500

type (
	msgBody map[string]any
)

var (
	neighbors []string
	// "set" of observed values by this node
	db utils.MapStruct[int, any]
	// map of [neighbor ID, set of values this node knows that the neighbor knows]
	// used for periodic Sync operations
	acked utils.MapStruct[string, map[int]any]
)

/*
 *  Request Handlers
 */

func handleBroadcast(n *maelstrom.Node, isSync bool) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		var ackBody msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		if err := json.Unmarshal(msg.Body, &ackBody); err != nil {
			return err
		}
		message := int(ackBody["message"].(float64))
		delete(ackBody, "message")
		ackBody["type"] = "broadcast_ok"
		if err := n.Reply(msg, ackBody); err != nil {
			return err
		}
		_, ok := db.Get(message)
		if ok {
			return nil // return if message is already handled
		}
		db.Put(message, nil)
		if isSync {
			return nil // return without propagation to neighbors if sync is used
		}
		waiting := sync.Map{} // use sync.Map for retried iterations
		for _, neighbor := range neighbors {
			if neighbor == msg.Src {
				continue
			}
			waiting.Store(neighbor, false)
		}
		pending := true
		var err error
		for pending {
			pending = false
			waiting.Range(func(neighbor, value any) bool {
				if v, _ := waiting.Load(neighbor); v.(bool) {
					return true
				}
				pending = true
				// RPCs are non-blocking, as opposed to SyncRPC
				err = n.RPC(neighbor.(string), body, func(msg maelstrom.Message) error {
					waiting.Store(neighbor, true)
					waiting.LoadOrStore(neighbor, true)
					return nil
				})
				return err == nil
			})
			if pending {
				// Each handler functions are run in its own goroutine, OK to sleep
				time.Sleep(time.Millisecond * RetryMs)
			}
		}
		return err
	}
}

func handleRead(n *maelstrom.Node) maelstrom.HandlerFunc {
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

func handleTopology(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "topology_ok"
		if UseTopologyMsg {
			if err := setNeighborFromMsg(n, &body); err != nil {
				return err
			}
		} else {
			setSpanningTreeNeighbor(n, NumChild)
		}
		// Set up the acked DS
		for _, id := range neighbors {
			acked.M[id] = make(map[int]any)
		}
		delete(body, "topology")
		return n.Reply(msg, body)
	}
}

func handleSync() maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// 1. Add newly seen messages to DB (& acked)
		received := body["message"].([]any)
		for _, v := range received {
			val := int(v.(float64))
			db.Put(val, nil)
			acked.Lock()
			curr := acked.M[msg.Src]
			curr[val] = nil
			acked.Unlock()
		}
		// 2. Send back entire DB - (acked by sender)
		var res []int
		// NOTE: Cannot use mapStruct.Get() here, since the values of the map
		// is returned as a reference, leading to concurrent map reads.
		ackedValues := make(map[int]any)
		acked.RLock()
		for val := range acked.M[msg.Src] {
			ackedValues[val] = nil
		}
		acked.RUnlock()
		for _, v := range *db.Keys() {
			if _, ok := ackedValues[v]; !ok {
				res = append(res, v)
			}
		}
		return nil
	}
}

/*
 *  Helpers
 */

func syncDB(n *maelstrom.Node) error {
	if n.ID() == "" {
		return nil
	}
	values := *db.Keys()
	body := make(msgBody)
	var message []int
	var currAcked map[int]any
	for _, neighbor := range neighbors {
		message = make([]int, 0)
		currAcked = make(map[int]any)
		acked.RLock()
		for val := range acked.M[neighbor] {
			currAcked[val] = nil
		}
		acked.RUnlock()
		for _, v := range values {
			if _, ok := currAcked[v]; !ok {
				message = append(message, v)
			}
		}
		body["type"] = "sync"
		body["message"] = message
		if err := n.Send(neighbor, &body); err != nil {
			return err
		}
	}
	return nil
}

func setNeighborFromMsg(n *maelstrom.Node, body *msgBody) error {
	field, ok := (*body)["topology"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid JSON body format: 'topology'")
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
	neighbors = topology[n.ID()]
	return nil
}

func setSpanningTreeNeighbor(n *maelstrom.Node, numChild int) {
	topology := make(map[string][]string)
	for _, node := range n.NodeIDs() {
		topology[node] = make([]string, 0)
	}
	queue := []int{0}
	idx := 1
	currNodeIdx := 0
	for len(queue) > 0 {
		currNodeIdx = queue[0]
		currNodeID := n.NodeIDs()[currNodeIdx]
		queue = queue[1:]
		for i := 0; i < numChild && idx < len(n.NodeIDs()); i++ {
			queue = append(queue, idx)
			neighborNodeID := n.NodeIDs()[idx]
			topology[currNodeID] = append(topology[currNodeID], neighborNodeID)
			topology[neighborNodeID] = append(topology[neighborNodeID], currNodeID)
			idx++
		}
	}
	neighbors = topology[n.ID()]
}

func main() {
	n := maelstrom.NewNode()
	db = utils.MapStruct[int, any]{
		RWMutex: sync.RWMutex{},
		M:       make(map[int]any),
	}
	acked = utils.MapStruct[string, map[int]any]{
		RWMutex: sync.RWMutex{},
		M:       make(map[string]map[int]any),
	}

	// Register handlers
	n.Handle("broadcast", handleBroadcast(n, UseSync))
	n.Handle("read", handleRead(n))
	n.Handle("topology", handleTopology(n))
	n.Handle("sync", handleSync())

	if UseSync {
		go func() {
			for {
				if err := syncDB(n); err != nil {
					log.Fatal(err)
				}
				time.Sleep(SyncMs * time.Millisecond)
			}
		}()
	}

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
