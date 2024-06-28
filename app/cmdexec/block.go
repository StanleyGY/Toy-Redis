package cmdexec

import (
	"github.com/stanleygy/toy-redis/app/algo"
)

const (
	BlockOnStream = 1
)

// Convert timeout unix timestamp into string and store them in radix tree.
// This makes it faster to search for clients that have timed out when there are
// many blocking clients.
var blockedClients *algo.RadixTree

func HandleBlockedClientsTimeout() {
	// iterating over all blocking clients
	// when blocking times out, unblock the client with a null reply
}

func BlockClientForKey(c *ClientInfo, source int, key string) {

}
