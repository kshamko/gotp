package actor

import (
	"fmt"

	"gopkg.in/tomb.v2"
)

/*const (
	//ReplyOK - ok status for actor response
	ReplyOK   = "ok"
	ReplyErr  = "error"
	ReplyDead = "dead"
)*/

//Reply is send by HandleCall or HandleCast
type Reply struct {
	Err      error
	Response interface{}
}

//StateInterface type is for internal actor state
type StateInterface interface{}

//Actor struct
type actor struct {
	messageChanSync  chan MessageInterface
	messageChanAsync chan MessageInterface
	replyChan        chan Reply
	state            StateInterface
	stateInitial     StateInterface
	messages         map[string]MessageInterface
	tmb              tomb.Tomb
}

//Start creates new actor
func (a *actor) start(state StateInterface, messages []MessageInterface) *Pid {
	a.state = state
	a.stateInitial = state
	a.messageChanSync = make(chan MessageInterface)
	a.messageChanAsync = make(chan MessageInterface)
	a.replyChan = make(chan Reply)
	a.messages = make(map[string]MessageInterface)

	for _, m := range messages {
		a.messages[m.GetType()] = m
	}

	a.tmb.Go(a.loop)

	return &Pid{
		a: a,
	}
}

//HandleCall makes sync actor call
func (a *actor) handleCall(message MessageInterface) Reply {
	//recover from panic while writing to chan when routineis dying
	defer func() { recover() }()
	if !a.tmb.Alive() {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}

	a.messageChanSync <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *actor) handleCast(message MessageInterface) Reply {
	//recover from panic while writing to chan when routineis dying
	defer func() { recover() }()
	if !a.tmb.Alive() {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}
	a.messageChanAsync <- message
	return Reply{nil, nil}
}

//Stop stops an actor
func (a *actor) stop() error {
	a.tmb.Kill(nil)
	return a.tmb.Wait()
}

//GetStopReason returns error led to actor stop
func (a *actor) getStopReason() error {
	return a.tmb.Err()
}

//main select loop
func (a *actor) loop() error {
	for {
		select {
		case <-a.tmb.Dying():
			a.closeAllChans()
			return a.tmb.Err()

		case msg := <-a.messageChanSync:
			reply := msg.Handle(a.state)
			a.replyChan <- reply.ActorReply
			if reply.Stop {
				return reply.Err
			}
			a.state = reply.State

		case msg := <-a.messageChanAsync:
			reply := msg.Handle(a.state)
			if reply.Stop {
				return reply.Err
			}
			a.state = reply.State
		}
	}
}

func (a *actor) closeAllChans() {
	close(a.messageChanAsync)
	close(a.messageChanSync)
	close(a.replyChan)
}
