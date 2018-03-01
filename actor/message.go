package actor

//
type messageInterface interface {
	Handle(StateInterface) MessageReply
}

//MessageReply is for reply by Handle()
type MessageReply struct {
	ActorReply Reply
	Stop       bool
	StopReason string
	State      StateInterface
}
