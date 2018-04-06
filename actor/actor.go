package actor

import (
	"errors"
)

var (
	//ErrRestarting - error sent by actory while it's restarting
	ErrRestarting = errors.New("actor: in restarting state")

	//ErrDead - error sent by dead actor
	ErrDead = errors.New("actor: dead")
)

//Reply is send by HandleCall or HandleCast
type Reply struct {
	Err      error
	Response interface{}
}

//StateInterface type is for internal actor state
type StateInterface interface{}

type actorInterface interface {
	start(init Initer, messages []MessageInterface) error
	init() error
	loop(readyChan chan bool) error
	setMonitor(m *monitor)
	getPid() pid

	HandleCall(message MessageInterface) Reply
	HandleCast(message MessageInterface) Reply
}

//Actor struct
type Actor struct {
	pid              pid
	messageChanSync  chan MessageInterface
	messageChanAsync chan MessageInterface
	replyChan        chan Reply
	state            StateInterface
	messages         map[string]struct{}
	initer           Initer
	spec             ChildSpec
	monitor          *monitor
	//dieChan    chan bool
	ready chan bool
}

//Initer defines function to set initial actor state
type Initer struct {
	Fn   func(params interface{}) (StateInterface, error)
	Args interface{}
}

// PUBLIC METHODS

//HandleCall makes sync actor call
func (a *Actor) HandleCall(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		return Reply{ErrRestarting, nil}
	}
	a.messageChanSync <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *Actor) HandleCast(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		return Reply{ErrRestarting, nil}
	}
	a.messageChanAsync <- message
	return Reply{nil, "ok"}
}

//WaitRestart should be used to wait actor to be restarted by supervisor after entering restart state
func (a *Actor) WaitRestart() error {
	return a.monitor.waitRestart()
}

// PRIVATE METHODS

//Start creates new actor
//TODO check that callbacks are defines. if not return
func (a *Actor) start(init Initer, messages []MessageInterface) error {

	a.pid = newPid()
	a.initer = init

	a.messages = make(map[string]struct{})
	for _, m := range messages {
		a.messages[m.GetType()] = struct{}{}
	}

	return a.init()
}

//
func (a *Actor) init() error {

	var err error
	a.state, err = a.initer.Fn(a.initer.Args)
	if err != nil {
		return err
	}

	a.messageChanSync = make(chan MessageInterface)
	a.messageChanAsync = make(chan MessageInterface)
	a.replyChan = make(chan Reply)

	a.ready = make(chan bool)
	go a.loop(a.ready)
	go func() {
		a.ready <- true
	}()

	return nil
}

//
func (a *Actor) setMonitor(m *monitor) {
	a.monitor = m
}

//
func (a *Actor) getPid() pid {
	return a.pid
}

//main select loop
func (a *Actor) loop(readyChan chan bool) error {
	for {
		select {
		case msg := <-a.messageChanSync:
			reply := msg.Handle(a.state)
			a.replyChan <- reply.ActorReply

			if reply.Stop {
				close(readyChan)
				return a.handleDie(reply.Err)
			}
			a.state = reply.State
			readyChan <- true

		case msg := <-a.messageChanAsync:
			reply := msg.Handle(a.state)

			if reply.Stop {
				close(readyChan)
				return a.handleDie(reply.Err)
			}
			a.state = reply.State
			readyChan <- true

		}
	}
}

func (a *Actor) handleDie(err error) error {
	close(a.messageChanAsync)
	close(a.messageChanSync)
	close(a.replyChan)
	if a.monitor != nil {
		a.monitor.trigger(err)
	}
	return err
}
