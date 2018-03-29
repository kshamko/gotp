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
}

type supChild struct {
	startTime  time.Time
	restarts   int
	actor      actorInterface
	supervisor bool
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

//
func (s *Sup) supervisorRestartChild(a *actor) error {
	r := s.HandleCall(msgRestartChild{a, s})
	return r.Err
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
	childStats := s.children[m.child.pid.id]
	if childStats.startTime.Add(2*m.child.spec.RestartRetryIn).UnixNano() < time.Now().UnixNano() {
		childStats.restarts = 0
	}

	childStats.restarts++

	if childStats.restarts > m.child.spec.RestartCount {
		delete(s.children, m.child.pid.id)

		return MessageReply{
			ActorReply: Reply{ErrDead, nil},
			State:      s,
		}
	}

	err := m.child.restart()
	if err != nil {
		return MessageReply{
			ActorReply: Reply{err, nil},
			State:      s,
		}
	}

	//setup monitor
	mon := newMonitor(m.sup, m.child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*actor).spec.RestartRetryIn)
		fmt.Println("call restart ", actr.(*actor).spec.RestartRetryIn)
		return sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})

	childStats.startTime = time.Now()
	s.children[m.child.pid.id] = childStats

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
	mon.start(func(sup, actr actorInterface) error {
		fmt.Println("call start")
		return sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})

	//add actor to supervisor state
	s.children[a.pid.id] = supChild{
		actor:     a,
		restarts:  0,
		startTime: time.Now(),
	}

	return MessageReply{
		ActorReply: Reply{nil, a},
		State:      s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}
