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
func (s *Sup) setupMonitor(child actorInterface) {
	mon := newMonitor(s, child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*actor).spec.RestartRetryIn)
		return sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})
}

func (s *Sup) checkChildStats(childInfo supChild) (int, error) {

	var err error

	restartCount := childInfo.restarts + 1
	restartThreshold := childInfo.actor.(*actor).spec.RestartRetryIn + 25*childInfo.actor.(*actor).spec.RestartRetryIn/100
	if childInfo.startTime.Add(restartThreshold).UnixNano() < time.Now().UnixNano() {
		restartCount = 1
	}

	if restartCount > childInfo.actor.(*actor).spec.RestartCount {
		err = ErrDead
	}

	return restartCount, err
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
	restartCount, err := m.sup.checkChildStats(childStats)

	if err != nil {
		delete(s.children, m.child.pid.id)

		return MessageReply{
			ActorReply: Reply{
				Err:      ErrDead,
				Response: nil,
			},
			State: s,
		}
	}

	err = m.child.restart()
	if err != nil {
		return MessageReply{
			ActorReply: Reply{
				Err:      err,
				Response: nil,
			},
			State: s,
		}
	}

	//update state
	childStats.startTime = time.Now()
	childStats.restarts = restartCount
	s.children[m.child.pid.id] = childStats

	//setup monitor
	m.sup.setupMonitor(m.child)

	return MessageReply{
		ActorReply: Reply{
			Err:      nil,
			Response: nil,
		},
		State: s,
	}
}

//
type msgStartChild struct {
	spec ChildSpec
	sup  *Sup
}

func (m msgStartChild) Handle(state StateInterface) MessageReply {

	a := &actor{
		spec: m.spec,
	}

	//start actor
	err := a.start(m.spec.Init, m.spec.Messages)
	s := state.(supState)
	if err != nil {
		return MessageReply{
			ActorReply: Reply{
				Err:      err,
				Response: nil,
			},
			State: s,
		}
	}

	//add actor to supervisor state
	s.children[a.pid.id] = supChild{
		actor:     a,
		restarts:  0,
		startTime: time.Now(),
	}

	//setup monitor
	m.sup.setupMonitor(a)

	return MessageReply{
		ActorReply: Reply{
			Err:      nil,
			Response: a,
		},
		State: s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}

// Private function
