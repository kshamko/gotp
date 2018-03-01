package actor

const (
	//ReplyOK - ok status for actor response
	ReplyOK = "ok"
)

//Actor struct
type Actor struct {
	messageChanSync  chan messageInterface
	messageChanAsync chan messageInterface
	replyChan        chan Reply
	exitChan         chan bool
	state            StateInterface
}

//Reply is send by HandleCall or HandleCast
type Reply struct {
	Err      error
	Response interface{}
}

//StateInterface type is for internal actor state
type StateInterface interface{}

//Start creates new actor
func (a *Actor) Start(state StateInterface) {
	a.state = state
	a.messageChanSync = make(chan messageInterface)
	a.messageChanAsync = make(chan messageInterface)
	a.replyChan = make(chan Reply)
	go a.loop()
}

//HandleCall makes sync actor call
func (a *Actor) HandleCall(message messageInterface) Reply {
	a.messageChanSync <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *Actor) HandleCast(message messageInterface) Reply {
	a.messageChanAsync <- message
	return Reply{nil, ReplyOK}
}

//Stop stops an actor
func (a *Actor) Stop() {
	a.exitChan <- true
}

//main select loop
func (a *Actor) loop() {
	for {
		select {
		case msg := <-a.messageChanSync:
			reply := msg.Handle(a.state)
			a.state = reply.State
			a.replyChan <- reply.ActorReply
			if reply.Stop {
				a.closeAllChans()
				return
			}
		case msg := <-a.messageChanAsync:
			reply := msg.Handle(a.state)
			a.state = reply.State
			if reply.Stop {
				a.closeAllChans()
				return
			}
		case <-a.exitChan:
			a.closeAllChans()
			return
		}
	}
}

func (a *Actor) closeAllChans() {
	close(a.messageChanAsync)
	close(a.messageChanSync)
	close(a.replyChan)
}
