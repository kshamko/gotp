package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/kshamko/gotp/actor"
)

type st struct {
	bal int
}

//messages
type addBalanceMsg struct {
	inc int
}

//
func (m addBalanceMsg) Handle(state actor.StateInterface) actor.MessageReply {
	st := state.(st)
	st.bal += m.inc
	if st.bal > 6783 {

		return actor.MessageReply{
			ActorReply: actor.Reply{
				//Err:     0 fmt.Errorf("Kill reply"),
				Response: st.bal,
			},
			Stop:  true,
			Err:   fmt.Errorf("Balance error"),
			State: st,
		}
	}

	return actor.MessageReply{
		State: st,
	}

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
	sup, _ := actor.SupervisorStart(actor.SupOneForOne)

	workerSpec := actor.ChildSpec{
		IsSupervisor:   true,
		RestartCount:   5,
		RestartRetryIn: 1 * time.Second,
		Init: actor.Initer{
			Fn: func(p interface{}) (actor.StateInterface, error) {
				return st{p.(int)}, nil
			},
			Args: 56,
		},
		Messages: []actor.MessageInterface{
			addBalanceMsg{},
			getBalanceMsg{},
		},
	}

	worker, _ := sup.SupervisorStartChild(workerSpec)

	p := sync.Mutex{}
	w := sync.WaitGroup{}
	msgs := 0

	for i := 0; i < 60000; i++ {
		w.Add(1)
		go func() {
			for j := 0; j < 20; j++ {
				p.Lock()
				msgs++
				p.Unlock()
				worker.HandleCast(addBalanceMsg{2})
			}
			w.Done()
		}()
		//time.Sleep(2 * time.Millisecond)
	}

	w.Wait()

	time.Sleep(2 * time.Second)
	fmt.Printf("Get Balance: %+v\n", worker.HandleCall(getBalanceMsg{}))
	fmt.Println("Msgs processed:", msgs)
	fmt.Println("Routines: ", runtime.NumGoroutine())
	fmt.Println(sup)
}
