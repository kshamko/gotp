package actor

import (
	"fmt"
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
	children map[string]supChild
	restarts int
}

type supChild struct {
	startTime time.Time
	restarts  int
	actor     actorInterface
}

//type actorData

//SupervisorStart starts a supervisor
func SupervisorStart(supType int) (*Sup, error) {
	msgs := []MessageInterface{
		msgStartChild{},
		msgRestartChild{},
	}
	s := &Sup{supType: supType}

	initer := Initer{
		Fn: func(p interface{}) (StateInterface, error) {
			return supState{
				children: make(map[string]supChild),
			}, nil
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

	//s.children[m.child.pid.id].restarts++

	childStats := s.children[m.child.pid.id]
	childStats.restarts++
	s.children[m.child.pid.id] = childStats
	fmt.Println("restarts ", childStats.restarts)

	if childStats.restarts > m.child.spec.RestartCount {
		//panic("ggg")
		return MessageReply{
			ActorReply: Reply{fmt.Errorf("actor_totally_dead"), nil},
			State:      s,
		}
	}

	//if m.child.spec.RestartRetryIn
	//afterChan := time.After(m.child.spec.RestartRetryIn)

	err := m.child.restart()
	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}

	s.restarts++
	mon := newMonitor(m.sup, m.child)
	//go func() {
	//<-time.Tick(m.child.spec.RestartRetryIn)

	mon.start(func(sup, actr actorInterface) {
		time.Sleep(m.child.spec.RestartRetryIn)
		sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})
	//}()

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

	a := &actor{}
	a.spec = m.spec

	err := a.start(m.spec.Init, m.spec.Messages)
	s := state.(supState)

	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}

	//setup monitor
	mon := newMonitor(m.sup, a)
	mon.start(func(sup, actr actorInterface) {
		//time.Sleep(actr.(*actor).spec.RestartRetryIn)
		sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})

	return MessageReply{
		ActorReply: Reply{nil, a},
		State:      s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}
