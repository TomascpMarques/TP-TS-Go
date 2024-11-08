package msgpacktyps

import "time"

type MessageType byte

const (
	RequestId   MessageType = 0
	SendContent             = iota
	RequestIdResponse
	SendContentResponse
)

type Message struct {
	Created  int64       `msgpack:"time"`
	SenderId string      `msgpack:"sender"`
	Type     MessageType `msgpack:"msg_type"`
	Target   string      `msgpack:"target"`
	Content  []byte      `msgpack:"content"`
}

func NewMessage(msgType MessageType, sender string, target string, content ...byte) Message {
	return Message{
		Created:  time.Now().UnixMilli(),
		Content:  content,
		Target:   target,
		Type:     msgType,
		SenderId: sender,
	}
}
