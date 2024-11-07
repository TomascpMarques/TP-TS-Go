package msgpacktyps

import "time"

type MessageType byte

const (
	RequestId   MessageType = iota + 1
	SendContent             = 0
)

type Message struct {
	Created time.Time   `msgpack:"time"`
	Type    MessageType `msgpack:"msg_type"`
	Target  string      `msgpack:"target"`
	Content []byte      `msgpack:"content"`
}

func NewMessage(target string, content ...byte) (msg Message) {
	msg.Created = time.Now()
	msg.Content = content
	msg.Target = target
	return
}
