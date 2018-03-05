package main

import (
	"fmt"
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
	if st.bal > 0 {

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
		IsSupervisor: true,
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

	worker, err := sup.SupervisorStartChild(workerSpec)

	fmt.Printf("WORKER_ACTOR: %v\n Err:%v\n\n", worker, err)

	m := sync.Mutex{}
	msgs := 0

	for i := 0; i < 40000; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				m.Lock()
				msgs += 1
				m.Unlock()

				worker.HandleCast(addBalanceMsg{1})

				//time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	/*for j := 0; j < 20; j++ {
		r := worker.HandleCast(addBalanceMsg{1})
		fmt.Println("Add", j, ":", r)
		//time.Sleep(50 * time.Millisecond)
	}*/
	//
	fmt.Printf("\n\n\nGet Balance: %+v\n", worker.HandleCall(getBalanceMsg{}))
	//a.Stop()
	time.Sleep(5 * time.Second)
	fmt.Printf("Get Balance: %+v, \n\n %+v\n", worker.HandleCall(getBalanceMsg{}), sup)
	fmt.Println(msgs)

	//a.Stop()
}
