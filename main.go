package main

import (
	"fmt"

	"github.com/kshamko/gotp/actor"
)

type st struct {
	bal int
}

type myActor struct {
	actor.Actor
}

func (a *myActor) AddBalance(x int) actor.Reply {
	return a.Actor.HandleCast(addBalanceMsg{x})
}

func (a *myActor) GetBalance() actor.Reply {
	reply := a.Actor.HandleCall(getBalanceMsg{})
	return reply
}

//messages
type addBalanceMsg struct {
	inc int
}

func (m addBalanceMsg) Handle(state interface{}) actor.Reply {
	st := state.(st)
	st.bal += m.inc
	return actor.Reply{Err: nil, Response: "ok", State: st}
}

type getBalanceMsg struct{}

func (m getBalanceMsg) Handle(state interface{}) actor.Reply {
	return actor.Reply{Err: nil, Response: state.(st).bal, State: state}
}

//
func main() {
	a := &myActor{}
	a.Start(st{0})

	for i := 0; i < 20; i++ {
		fmt.Printf("Add balance %+v\n", a.AddBalance(7))
	}

	fmt.Printf("Get Balance: %+v\n", a.GetBalance())
}
