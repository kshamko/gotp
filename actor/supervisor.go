package actor

import (
	"sync"
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
	m       sync.Mutex
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

//SupervisorStart
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
	a *actor
	s *sup
}

func (m msgRestartChild) GetType() string {
	return "restart_child"
}

func (m msgRestartChild) Handle(state StateInterface) MessageReply {
	dieChan := make(chan bool)

	messages := m.a.messages
	newa, _ := m.a.start(m.a.initer, []MessageInterface{})

	newa.messages = messages
	newa.dieChan = dieChan
	m.a = nil

	s := state.(supState)
	s.restarts++

	go func() {
		restart := <-dieChan
		if restart {
			m.s.supervisorRestartChild(newa)
		}
		close(dieChan)
	}()

	return MessageReply{
		ActorReply: Reply{nil, newa},
		State:      s,
	}

}

type msgStartChild struct {
	spec ChildSpec
	sup  *sup
}

func (m msgStartChild) Handle(state StateInterface) MessageReply {

	dieChan := make(chan bool)
	a := &actor{
		dieChan: dieChan,
	}
	actor, err := a.start(m.spec.Init, m.spec.Messages)

	s := state.(supState)

	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}
	//s.children = append(s.children, a)

	actor.supervisor = m.sup
	actor.spec = m.spec

	go func() {
		restart := <-dieChan
		if restart {
			m.sup.supervisorRestartChild(a)
		}
		close(dieChan)
	}()

	return MessageReply{
		ActorReply: Reply{nil, actor},
		State:      s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}
