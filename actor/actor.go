package actor

import (
	"fmt"
	"sync"
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

//func Start(a ActorInterface)

//Actor struct
type actor struct {
	pid              pid
	messageChanSync  chan MessageInterface
	messageChanAsync chan MessageInterface
	replyChan        chan Reply
	state            StateInterface
	messages         map[string]struct{}

	initer Initer

	supervisor *sup
	spec       ChildSpec
	dieChan    chan bool

	ready chan bool
}

//
type Initer struct {
	Fn   func(params interface{}) (StateInterface, error)
	Args interface{}
}

//Start creates new actor
//TODO check that callbacks are defines. if not return
func (a *actor) start(init Initer, messages []MessageInterface) (*actor, error) {

	var err error
	a.state, err = init.Fn(init.Args)
	if err != nil {
		return nil, err
	}

	a.messageChanSync = make(chan MessageInterface)
	a.messageChanAsync = make(chan MessageInterface)
	a.replyChan = make(chan Reply)

	a.pid = newPid()
	a.initer = init
	a.messages = make(map[string]struct{})
	for _, m := range messages {
		a.messages[m.GetType()] = struct{}{}
	}

	//
	stopCh := make(chan error, 1)
	go func() {
		stopErr := <-stopCh
		if stopErr != nil {
			a.handelDie(stopErr)
		}
		close(stopCh)
	}()

	//set actor ready
	a.ready = make(chan bool)
	go a.loop(stopCh, a.ready)
	go func() {
		a.ready <- true
	}()
	return a, nil
}

//HandleCall makes sync actor call
func (a *actor) HandleCall(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}
	a.messageChanSync <- message
	return <-a.replyChan
}

var (
	ERR, OK int
	m       = &sync.RWMutex{}
)

//HandleCast makes async call to actor
func (a *actor) HandleCast(message MessageInterface) Reply {
	if ready := <-a.ready; !ready {
		m.Lock()
		ERR++
		m.Unlock()
		return Reply{fmt.Errorf("actor_dead"), nil}
	}
	m.Lock()
	OK++
	m.Unlock()
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

//main select loop
func (a *actor) loop(stopChan chan error, readyChan chan bool) error {
	for {

		select {

		case msg := <-a.messageChanSync:
			reply := msg.Handle(a.state)
			a.replyChan <- reply.ActorReply
			a.state = reply.State

			if reply.Stop {
				close(readyChan)
				a.handelDie(reply.Err)
				return reply.Err
			}

			readyChan <- true
		case msg := <-a.messageChanAsync:
			reply := msg.Handle(a.state)
			a.state = reply.State

			if reply.Stop {
				close(readyChan)
				a.handelDie(reply.Err)
				return reply.Err
			}
			readyChan <- true
		}

	}
}

var RS int

func (a *actor) handelDie(err error) {
	close(a.messageChanAsync)
	close(a.messageChanSync)
	close(a.replyChan)

	if err != nil {
		RS++
		a.dieChan <- true
	}
}
