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

var blockClients map[int]string
var blockClientsOnKeySpace map[BlockKey][]*ClientInfo
var pendingClients []*ClientInfo

func MakeBlockList() {
	clientsTimeoutTable = algo.MakeRadixTree()
	blockClients = make(map[int]string)
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
		// Send a null reply when a timeout occurs
		AddNullBulkStringReplyEvent(c)
		UnblockClient(c)
		log.Println("A client timeout occurs:", c.ConnFd)
	}
}

func GetEarliestTimeoutUnix() int {
	timeoutStartId := strconv.FormatInt(time.Now().UnixMilli(), 10)
	searchResults := clientsTimeoutTable.SearchByRange(timeoutStartId, ":", 1)
	if len(searchResults) == 0 {
		return -1
	}

	// Calculate time elapsed between now and the client timeout
	r := searchResults[0]
	timeoutMs, err := strconv.ParseInt(r.Id, 10, 64)
	if err != nil {
		log.Println(err.Error())
		return -1
	}
	nowMs := time.Now().UnixMilli()
	return int(timeoutMs - nowMs)
}

func ReprocessPendingClients() {
	for _, c := range pendingClients {
		Execute(c, c.ClientRequest)
		log.Println("A client is reprocessed:", c.ConnFd)
	}
	pendingClients = make([]*ClientInfo, 0)
}

func UnblockClient(c *ClientInfo) {
	timeoutId, found := blockClients[c.ConnFd]
	if found {
		delete(blockClients, c.ConnFd)
		clientsTimeoutTable.Remove(timeoutId)
	}
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

	// Add client to key space block list
	_, found = blockClientsOnKeySpace[*bkey]
	if !found {
		blockClientsOnKeySpace[*bkey] = []*ClientInfo{}
	}
	blockClientsOnKeySpace[*bkey] = append(blockClientsOnKeySpace[*bkey], c)

	// Add client to timeout table
	timeoutId := strconv.FormatInt(time.Now().Add(time.Millisecond*time.Duration(timeoutMs)).UnixMilli(), 10)
	blockClients[c.ConnFd] = timeoutId
	clientsTimeoutTable.Insert(timeoutId, c)
}
