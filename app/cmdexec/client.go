package cmdexec

import "github.com/stanleygy/toy-redis/app/resp"

type ClientInfo struct {
	ConnFd        int
	ClientRequest *resp.RespValue
}
