package actor

//MessageInterface is for messages received and processed by actors
type MessageInterface interface {
	Handle(StateInterface) MessageReply
	GetType() string
}

//MessageReply is for reply by Handle()
type MessageReply struct {
	ActorReply Reply
	Stop       bool
	Err        error
	State      StateInterface
}
