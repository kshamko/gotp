package main

import (
	"fmt"
	"time"

	"github.com/kshamko/gotp/actor"
)

type st struct {
	bal int
}

/*
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
*/

//messages
type addBalanceMsg struct {
	inc int
}

//
func (m addBalanceMsg) Handle(state actor.StateInterface) actor.MessageReply {
	st := state.(st)
	if st.bal > 400 {
		return actor.MessageReply{
			/*Stop:       true,
			StopReason: fmt.Errorf("Balance error"),*/
			Err:   actor.HandleError{true, fmt.Errorf("Balance error")},
			State: st,
		}
	}

	st.bal += m.inc

	return actor.MessageReply{State: st}
}

func (m addBalanceMsg) GetType() string {
	return "add_balance"
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

func (m getBalanceMsg) GetType() string {
	return "get_balance"
}

//
func main() {
	sup := actor.SupervisorStart(actor.SupOneForOne)

	workerSpec := actor.ChildSpec{
		IsSupervisor: true,
		State:        st{},
		Messages: []actor.MessageInterface{
			addBalanceMsg{},
			getBalanceMsg{},
		},
	}

	worker, _ := actor.SupervisorStartChild(sup, workerSpec)

	fmt.Printf("Sup: %+v\n", sup)
	fmt.Printf("Wrk: %+v\n", worker)

	for i := 0; i < 40000; i++ {
		go func() {
			for i := 0; i < 20; i++ {
				worker.Cast(addBalanceMsg{7}) //AddBalance(7)
			}
		}()
	}

	fmt.Printf("Get Balance: %+v\n", worker.Call(getBalanceMsg{}))
	//a.Stop()
	time.Sleep(2 * time.Second)
	fmt.Printf("Get Balance: %+v, %v\n", worker.Call(getBalanceMsg{}), worker.GetStopReason())
}
