package actor

type actorMessageInterface interface {
	Handle(interface{}) Reply
}

type Reply struct {
	Err error
	//Stop     bool
	Response interface{}
	State    interface{}
}

type StateInterface interface {
	getState() StateInterface
	setState(StateInterface)
}

//Actor struct
type Actor struct {
	messageChan chan actorMessageInterface
	replyChan   chan Reply
	exitChan    chan bool
	state       interface{}
	respond     bool
}

//Start creates new actor
func (a *Actor) Start(state interface{}) {
	a.state = state
	a.messageChan = make(chan actorMessageInterface, 2)

	replyChan := make(chan Reply)
	a.replyChan = replyChan

	go a.loop()
}

//HandleCall makes sync actor call
func (a *Actor) HandleCall(message actorMessageInterface) Reply {
	a.respond = true
	a.messageChan <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *Actor) HandleCast(message actorMessageInterface) Reply {
	a.messageChan <- message
	return Reply{nil, "ok", a.state}
}

//private function

//main select loop
func (a *Actor) loop() {
	for {
		select {
		case msg := <-a.messageChan:

			reply := msg.Handle(a.state)
			a.state = reply.State

			if a.respond {
				a.replyChan <- reply
				a.respond = false
			}

		case <-a.exitChan:
			return
		}
	}
}
