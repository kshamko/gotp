package actor

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActorUnderLoad(t *testing.T) {

	sup, _ := SupervisorStart(actor.SupOneForOne)
	worker, _ := sup.SupervisorStartChild(getWorketSpec())

	w := sync.WaitGroup{}

	for i := 0; i < 40000; i++ {
		w.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				worker.HandleCast(addBalanceMsg{1})
			}
			w.Done()
		}()
	}

	w.Wait()

	assert.Equal(t, 4000000, actor.HandleCall(getBalanceMsg{}))
}

func getWorketSpec() ChildSpec {
	return ChildSpec{
		IsSupervisor:   false,
		RestartCount:   3,
		RestartRetryIn: 100 * time.Millisecond,
		Init: Initer{
			Fn: func(p interface{}) (StateInterface, error) {
				return testChildState{p.(int)}, nil
			},
			Args: 0,
		},
		Messages: []MessageInterface{
			addBalanceMsg{},
			getBalanceMsg{},
		},
	}

}

//worker definition
type testChildState struct {
	bal int
}
type addBalanceMsg struct {
	inc int
}

//
func (m addBalanceMsg) Handle(state StateInterface) MessageReply {
	st := state.(testChildState)
	st.bal += m.inc

	return MessageReply{
		State: st,
	}

}

func (m addBalanceMsg) GetType() string {
	return "add_balance"
}

//
type getBalanceMsg struct{}

func (m getBalanceMsg) Handle(state StateInterface) MessageReply {

	return MessageReply{
		ActorReply: Reply{
			Response: state.(testChildState).bal,
		},
		State: state,
	}
}

func (m getBalanceMsg) GetType() string {
	return "get_balance"
}
