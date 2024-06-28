package cmdexec

import "github.com/stanleygy/toy-redis/app/algo"

// Convert timeout unix timestamp into string and store them in radix tree.
// This makes it faster to search for clients that have timed out when there are
// many blocking clients.
var clientsTimeoutTable *algo.RadixTree

func HandleBlockedClientsTimeout() {
	// iterating over all blocking clients
	// when blocking times out, unblock the cllient
}
