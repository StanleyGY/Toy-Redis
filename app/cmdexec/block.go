package cmdexec

import (
	"log"
	"math"
	"strconv"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
)

const (
	BlockOnStream = 1
)

type BlockKey struct {
	Source int
	Key    string
}

// Convert timeout unix timestamp into string and store them in radix tree.
// This makes it faster to search for clients that have timed out when there are
// many blocking clients.
var clientsTimeoutTable *algo.RadixTree

var blockClients map[int]bool
var blockClientsOnKeySpace map[BlockKey][]*ClientInfo
var pendingClients []*ClientInfo

func MakeBlockList() {
	clientsTimeoutTable = algo.MakeRadixTree()
	blockClients = make(map[int]bool)
	blockClientsOnKeySpace = make(map[BlockKey][]*ClientInfo)
	pendingClients = make([]*ClientInfo, 0)
}

func HandleBlockedClientsTimeout() {
	// iterating over all blocking clients
	// when blocking times out, unblock the client with a null reply
	unixMs := strconv.FormatInt(time.Now().UnixMilli(), 10)

	searchResults := clientsTimeoutTable.SearchByRange("0", unixMs, math.MaxInt)
	for _, r := range searchResults {
		// Check if client is still being blocked
		c := r.Node.Value.(*ClientInfo)
		if blockClients[c.ConnFd] {
			// Send a null reply if a timeout occurs
			AddNullBulkStringReplyEvent(c)
			log.Println("A client timeout occurs:", c.ConnFd)
		}
		// Remove from timeout table
		clientsTimeoutTable.Remove(r.Id)
	}
}

func ReprocessPendingClients() {
	for _, c := range pendingClients {
		Execute(c, c.ClientRequest)
		log.Println("A client is reprocessed:", c.ConnFd)
	}
	pendingClients = make([]*ClientInfo, 0)
}

func UnblockClient(c *ClientInfo) {
	delete(blockClients, c.ConnFd)
}

func NotifyBlockedClientsOnKeySpace(bkey *BlockKey) {
	keyBlockList, found := blockClientsOnKeySpace[*bkey]
	if !found {
		return
	}
	// Remove clients from key space block list
	pendingClients = append(pendingClients, keyBlockList...)
	delete(blockClientsOnKeySpace, *bkey)
}

func BlockClientForKey(c *ClientInfo, bkey *BlockKey, timeoutMs int) {
	// Make sure each client is blocked only once
	_, found := blockClients[c.ConnFd]
	if found {
		return
	}
	blockClients[c.ConnFd] = true

	// Add client to key space block list
	_, found = blockClientsOnKeySpace[*bkey]
	if !found {
		blockClientsOnKeySpace[*bkey] = []*ClientInfo{}
	}
	blockClientsOnKeySpace[*bkey] = append(blockClientsOnKeySpace[*bkey], c)

	// Add client to timeout table
	unixMs := strconv.FormatInt(time.Now().Add(time.Millisecond*time.Duration(timeoutMs)).UnixMilli(), 10)
	clientsTimeoutTable.Insert(unixMs, c)
}
