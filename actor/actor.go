package actor

import (
	"fmt"
	"sync"

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

//func Start(a ActorInterface)

//Actor struct
type actor struct {
	pid              pid
	messageChanSync  chan MessageInterface
	messageChanAsync chan MessageInterface
	replyChan        chan Reply
	state            StateInterface
	tmb              *tomb.Tomb
	messages         map[string]struct{}

	initer Initer

	supervisor *sup
	spec       ChildSpec
	dieChan    chan bool
	isAlive    bool
	m          *sync.RWMutex
}

//
type Initer struct {
	Fn   func(params interface{}) (StateInterface, error)
	Args interface{}
}

//Start creates new actor
//TODO check that callbacks are defines. if not return
func (a *actor) start(init Initer, messages []MessageInterface) (*actor, error) {

	a.m = &sync.RWMutex{}
	var err error

	a.m.Lock()
	a.state, err = init.Fn(init.Args)
	if err != nil {
		return nil, err
	}
	a.messageChanSync = make(chan MessageInterface)
	a.messageChanAsync = make(chan MessageInterface, 1000)
	a.replyChan = make(chan Reply)

	a.tmb = &tomb.Tomb{}

	a.messages = make(map[string]struct{})
	a.pid = newPid()
	for _, m := range messages {
		a.messages[m.GetType()] = struct{}{}
	}

	a.initer = init
	a.tmb.Go(a.loop)

	go func() error {
		select {
		case <-a.tmb.Dying():

			//a.isAlive = false
			//err := a.tmb.Wait()
			//close(a.messageChanAsync)
			//close(a.messageChanSync)
			//close(a.replyChan)
			//return nil

			a.handelDie(true)
			return a.tmb.Wait()
		}
	}()

	a.isAlive = true
	a.m.Unlock()
	return a, nil
}

//HandleCall makes sync actor call
func (a *actor) HandleCall(message MessageInterface) Reply {
	//recover from panic while writing to chan when routineis dying
	defer func() {
		recover()
		//return Reply{fmt.Errorf("actor_dead"), nil}
	}()
	if !a.isAlive {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}

	a.messageChanSync <- message
	return <-a.replyChan
}

//HandleCast makes async call to actor
func (a *actor) HandleCast(message MessageInterface) Reply {
	//recover from panic while writing to chan when routineis dying
	defer func() {
		recover()
		//return Reply{fmt.Errorf("actor_dead"), nil}
	}()
	if !a.isAlive {
		return Reply{fmt.Errorf("actor_dead"), nil}
	}

	//fmt.Println(a.isAlive, a.tmb.Alive())
	a.messageChanAsync <- message

	return Reply{nil, "ok"}
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
		case msg := <-a.messageChanSync:
			//if msg != nil {
			//if a.isAlive {
			reply := msg.Handle(a.state)
			a.replyChan <- reply.ActorReply
			if reply.Stop {
				//a.tmb.Kill(reply.Err)
				//a.isAlive = false
				//a.dieChan <- true
				//a.handelDie(true)
				return reply.Err
			}
			a.state = reply.State
			//}
		case msg := <-a.messageChanAsync:
			//if msg != nil {
			//if a.isAlive {
			reply := msg.Handle(a.state)
			if reply.Stop {
				//a.tmb.Kill(reply.Err)
				//a.isAlive = false
				//a.dieChan <- true
				//a.handelDie(true)
				return reply.Err
			}
			a.state = reply.State
			//}
		}
	}
}

func (a *actor) handelDie(restart bool) {
	//defer func() { recover() }()
	//if !a.isAlive {
	//fmt.Println("HANDLE KILL")
	//panic("sss")
	//a.supervisor.supervisorRestartChild(a)
	a.dieChan <- true
	if a.isAlive {
		close(a.messageChanAsync)
		close(a.messageChanSync)
		close(a.replyChan)
	}
	a.isAlive = false

	//}
}
