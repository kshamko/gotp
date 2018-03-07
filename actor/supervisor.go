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
type Sup struct {
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

//SupervisorStart starts a supervisor
func SupervisorStart(supType int) (*Sup, error) {
	msgs := []MessageInterface{
		msgStartChild{},
	}
	s := &Sup{supType: supType}

	initer := Initer{
		Fn: func(p interface{}) (StateInterface, error) {
			return supState{}, nil
		},
	}

	s.start(initer, msgs)
	return s, nil
}

//SupervisorStartChild starts child
func (s *Sup) SupervisorStartChild(spec ChildSpec) (*actor, error) {
	res := s.HandleCall(msgStartChild{spec, s})
	return res.Response.(*actor), nil
}

func (s *Sup) supervisorRestartChild(a *actor) error {
	s.HandleCall(msgRestartChild{a, s})
	return nil
}

//
// SUPERVISOR ACTOR MESSAGES
//

type msgRestartChild struct {
	child *actor
	sup   *Sup
}

func (m msgRestartChild) GetType() string {
	return "restart_child"
}

func (m msgRestartChild) Handle(state StateInterface) MessageReply {

	s := state.(supState)

	err := m.child.restart()
	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}

	s.restarts++

	go func() {
		restart := <-m.child.dieChan
		if restart {
			m.sup.supervisorRestartChild(m.child)
		}
	}()

	return MessageReply{
		ActorReply: Reply{nil, nil},
		State:      s,
	}
}

//
type msgStartChild struct {
	spec ChildSpec
	sup  *Sup
}

func (m msgStartChild) Handle(state StateInterface) MessageReply {

	actor := &actor{}
	err := actor.start(m.spec.Init, m.spec.Messages)

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
