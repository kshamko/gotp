package actor

import (
	"time"
)

const (
	//SupOneForOne - restart strategy means that  your supervisor supervises
	//many workers and one of them fails, only that one should be restarted
	SupOneForOne = iota

	//SupOneForAll - restart strategy means that if a child process terminates,
	//all other child processes are terminated, and then all child processes,
	//including the terminated one, are restarted.
	SupOneForAll

	//SupRestForAll - restart strategy will terminate all actors started after
	//terminated one. Then the terminated child process and the rest of the child
	//processes are restarted.
	SupRestForAll

	//SupSimpleOneForOne
)

//Sup represents supervisor
type Sup struct {
	Actor
	supType int
}

// ChildSpec - specification of child which contains Init functions,
// IsSupervisor flag (if child is another supervisor), restart intencity (RestartCount
// happened in )
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
	startTime time.Time
	restarts  int
	actor     actorInterface
	//	supervisor bool
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

	err := s.start(initer, msgs)
	return s, err
}

//StartChild starts child
func (s *Sup) StartChild(spec ChildSpec) (*Actor, error) {
	res := s.HandleCall(msgStartChild{spec, s})
	return res.Response.(*Actor), nil
}

//TerminateChild gracefully stops child
func (s *Sup) TerminateChild() {

}

//
func (s *Sup) restartChild(a *Actor) error {
	r := s.HandleCall(msgRestartChild{a, s})
	return r.Err
}

//
func (s *Sup) setupMonitor(child actorInterface) {

	var restarter restartPolicy

	if s.supType == SupOneForOne {
		restarter = policyOneForOne{}
	}

	restarter.setupMonitor(s, child)
}

func (s *Sup) checkChildStats(childInfo supChild) (int, error) {

	var err error

	restartCount := childInfo.restarts + 1
	restartThreshold := childInfo.actor.(*Actor).spec.RestartRetryIn + 25*childInfo.actor.(*Actor).spec.RestartRetryIn/100
	if childInfo.startTime.Add(restartThreshold).UnixNano() < time.Now().UnixNano() {
		restartCount = 1
	}

	if restartCount > childInfo.actor.(*Actor).spec.RestartCount {
		err = ErrDead
	}

	return restartCount, err
}

//
// SUPERVISOR ACTOR MESSAGES
//

//
type msgRestartChild struct {
	child *Actor
	sup   *Sup
}

//
func (m msgRestartChild) GetType() string {
	return "restart_child"
}

//
func (m msgRestartChild) Handle(state StateInterface) MessageReply {

	s := state.(supState)
	childStats := s.children[m.child.pid.id]
	restartCount, err := m.sup.checkChildStats(childStats)

	//restart error
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

	//setup monitor
	m.sup.setupMonitor(m.child)

	//init actor
	err = m.child.init()
	if err != nil {
		m.child.monitor.trigger(err)
		return MessageReply{
			ActorReply: Reply{
				Err:      err,
				Response: nil,
			},
			State: s,
		}
	}

	//update sup state
	childStats.startTime = time.Now()
	childStats.restarts = restartCount
	s.children[m.child.pid.id] = childStats

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

//
func (m msgStartChild) GetType() string {
	return "add_child"
}

//
func (m msgStartChild) Handle(state StateInterface) MessageReply {

	a := &Actor{
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

// Private function
