package actor

import (
	"time"
)

const (
	SupOneForOne = iota
	SupOneForAll
	SupRestForAll
	SupSimpleOneForOne
)

//Sup
type sup struct {
	actor
	supType int
}

type ChildSpec struct {
	Init           Initer
	IsSupervisor   bool
	RestartCount   int
	RestartRetryIn time.Duration
	Messages       []MessageInterface
}

type supState struct {
	children []*actor
	restarts int
}

//type actorData

//SupervisorStart starta a supervisor
func SupervisorStart(supType int) (*sup, error) {
	msgs := []MessageInterface{
		msgStartChild{},
	}
	s := &sup{supType: supType}

	initer := Initer{
		Fn: func(p interface{}) (StateInterface, error) {
			return supState{}, nil
		},
	}

	s.start(initer, msgs)
	return s, nil
}

//SupervisorStartChild starts child
func (s *sup) SupervisorStartChild(spec ChildSpec) (*actor, error) {
	res := s.HandleCall(msgStartChild{spec, s})
	return res.Response.(*actor), nil
}

func (s *sup) supervisorRestartChild(a *actor) error {
	s.HandleCall(msgRestartChild{a, s})
	return nil
}

type msgRestartChild struct {
	child *actor
	sup   *sup
}

func (m msgRestartChild) GetType() string {
	return "restart_child"
}

func (m msgRestartChild) Handle(state StateInterface) MessageReply {
	messages := m.child.messages
	newa, _ := m.child.start(m.child.initer, []MessageInterface{})
	newa.messages = messages

	s := state.(supState)
	s.restarts++

	go func() {
		restart := <-newa.dieChan
		if restart {
			m.sup.supervisorRestartChild(newa)
		}
	}()

	return MessageReply{
		ActorReply: Reply{nil, newa},
		State:      s,
	}

}

//
type msgStartChild struct {
	spec ChildSpec
	sup  *sup
}

func (m msgStartChild) Handle(state StateInterface) MessageReply {

	a := &actor{}
	actor, err := a.start(m.spec.Init, m.spec.Messages)

	s := state.(supState)

	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}

	actor.supervisor = m.sup
	actor.spec = m.spec

	go func() {
		restart := <-actor.dieChan
		if restart {
			m.sup.supervisorRestartChild(actor)
		}
	}()

	return MessageReply{
		ActorReply: Reply{nil, actor},
		State:      s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}
