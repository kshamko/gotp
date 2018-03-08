package actor

import (
	"fmt"
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
	restart() error
	HandleCall(message MessageInterface) Reply
	HandleCast(message MessageInterface) Reply
	loop(readyChan chan bool) error
	stop() error
	setMonitor(m *monitor)
}

//Actor struct
type actor struct {
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

//
type Initer struct {
	Fn   func(params interface{}) (StateInterface, error)
	Args interface{}
}

//Start creates new actor
//TODO check that callbacks are defines. if not return
func (a *actor) start(init Initer, messages []MessageInterface) error {

	var err error
	a.state, err = init.Fn(init.Args)
	if err != nil {
		return err
	}
	a.pid = newPid()
	a.initer = init

	a.messages = make(map[string]struct{})
	for _, m := range messages {
		a.messages[m.GetType()] = struct{}{}
	}

	return a.restart()
}

func (a *actor) restart() error {
	a.messageChanSync = make(chan MessageInterface)
	a.messageChanAsync = make(chan MessageInterface)
	a.replyChan = make(chan Reply)
	//a.dieChan = make(chan bool)
	a.ready = make(chan bool)
	go a.loop(a.ready)
	go func() {
		a.ready <- true
		return
	}()

	return nil
}

//HandleCall makes sync actor call
func (a *actor) HandleCall(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}
	a.messageChanSync <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *actor) HandleCast(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}
	a.messageChanAsync <- message
	return Reply{nil, "ok"}
}

//Stop stops an actor
func (a *actor) stop() error {
	return nil
}

//GetStopReason returns error led to actor stop
func (a *actor) getStopReason() error {
	return nil
}

//GetStopReason returns error led to actor stop
func (a *actor) setMonitor(m *monitor) {
	a.monitor = m
	//return nil
}

//main select loop
func (a *actor) loop(readyChan chan bool) error {
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

func (a *actor) handleDie(err error) error {
	close(a.messageChanAsync)
	close(a.messageChanSync)
	close(a.replyChan)
	//if err != nil {
	a.monitor.trigger(err)
	//a.dieChan <- true
	//}
	return err
}
