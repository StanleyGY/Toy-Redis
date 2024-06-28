package cmdexec

import "github.com/stanleygy/toy-redis/app/resp"

const (
	EventReplyToClient = 1
)

var EventBus []*Event

type Event struct {
	Type   int
	Client *ClientInfo
	Resp   *resp.RespValue
	BKey   *BlockKey
}

func MakeEventBus() {
	EventBus = make([]*Event, 0)
}

func Reset() {
	MakeEventBus()
}

func AddEvent(ev *Event) {
	EventBus = append(EventBus, ev)
}

func AddReplyEvent(c *ClientInfo, r *resp.RespValue) {
	AddEvent(&Event{
		Type:   EventReplyToClient,
		Client: c,
		Resp:   r,
	})
}

func AddSimpleStringReplyEvent(c *ClientInfo, msg string) {
	AddReplyEvent(c, &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: msg})
}

func AddBulkStringReplyEvent(c *ClientInfo, msg string) {
	AddReplyEvent(c, resp.MakeBulkString(msg))
}

func AddNullBulkStringReplyEvent(c *ClientInfo) {
	AddReplyEvent(c, &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true})
}

func AddIntegerReplyEvent(c *ClientInfo, v int) {
	AddReplyEvent(c, resp.MakeInt(v))
}

func AddErrorReplyEvent(c *ClientInfo, err error) {
	AddReplyEvent(c, resp.MakeErorr(err.Error()))
}

func AddArrayReplyEvent(c *ClientInfo, arr []*resp.RespValue) {
	AddReplyEvent(c, &resp.RespValue{DataType: resp.TypeArrays, Array: arr})
}
