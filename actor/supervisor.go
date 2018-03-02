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
	State          StateInterface
	IsSupervisor   bool
	RestartCount   int
	RestartRetryIn time.Duration
	Messages       []MessageInterface
}

type supState struct {
	children []*actor
}

//SupervisorStart
func SupervisorStart(supType int) *Pid {
	msgs := []MessageInterface{
		msgStartChild{},
	}
	s := &sup{supType: supType}
	return s.start(supState{}, msgs)
}

//SupervisorStartChild starts child
func SupervisorStartChild(supPid *Pid, spec ChildSpec) (*Pid, error) {
	res := supPid.Call(msgStartChild{spec})
	return res.Response.(*Pid), nil
}

/*func (sup *Sup) StopChild() {

}

func (sup *Sup) restartChild() {

}*/

//////////
type msgStartChild struct {
	spec ChildSpec
}

func (m msgStartChild) Handle(state StateInterface) MessageReply {
	a := &actor{}
	pid := a.start(m.spec.State, m.spec.Messages)
	s := state.(supState)
	s.children = append(s.children, a)

	return MessageReply{
		ActorReply: Reply{nil, pid},
		State:      s,
	}
}

func (m msgStartChild) GetType() string {
	return "add_child"
}
