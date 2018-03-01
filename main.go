package main

import (
	"fmt"
	"time"

	"github.com/kshamko/gotp/actor"
)

type st struct {
	bal int
}

type myActor struct {
	actor.Actor
}

func (a *myActor) AddBalance(x int) actor.Reply {
	return a.HandleCast(addBalanceMsg{x})
}

func (a *myActor) GetBalance() actor.Reply {
	reply := a.HandleCall(getBalanceMsg{})
	return reply
}

//messages
type addBalanceMsg struct {
	inc int
}

//
func (m addBalanceMsg) Handle(state actor.StateInterface) actor.MessageReply {
	st := state.(st)
	st.bal += m.inc
	return actor.MessageReply{State: st}
}

type getBalanceMsg struct{}

func (m getBalanceMsg) Handle(state actor.StateInterface) actor.MessageReply {

	return actor.MessageReply{
		ActorReply: actor.Reply{
			Response: state.(st).bal,
		},
		State: state,
	}
}

//
func main() {
	a := &myActor{}
	a.Start(st{0})

	for i := 0; i < 40000; i++ {
		go func() {
			for i := 0; i < 20; i++ {
				a.AddBalance(7)
			}
		}()
	}

	fmt.Printf("Get Balance: %+v\n", a.GetBalance())
	time.Sleep(2 * time.Second)
	fmt.Printf("Get Balance: %+v\n", a.GetBalance())
}
