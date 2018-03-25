//go test -cover -coverprofile=c.out ./actor
//go tool cover -html=c.out

package actor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupervisorStart(t *testing.T) {
	sup, err := SupervisorStart(SupOneForAll)
	assert.Nil(t, err)
	assert.NotEmpty(t, sup.pid.id)
	assert.Equal(t, 0, len(sup.state.(supState).children))
}

func TestSupervisorStartChild(t *testing.T) {
	sup, _ := SupervisorStart(SupOneForAll)

	worker, err := sup.SupervisorStartChild(getWorketSpec())

	assert.Nil(t, err)
	assert.NotEmpty(t, worker.pid.id)
	assert.Equal(t, 1, len(sup.state.(supState).children))
}

func TestSupervisorRestartChild(t *testing.T) {
	sup, _ := SupervisorStart(SupOneForAll)
	worker, _ := sup.SupervisorStartChild(getWorketSpec())
	assert.Equal(t, 0, sup.state.(supState).children[worker.pid.id].restarts)

	res := worker.HandleCall(addBalanceMsgWithErr{2})
	worker.HandleCall(addBalanceMsgWithErr{2})
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)

	res1 := worker.HandleCall(addBalanceMsgWithErr{5})
	assert.Equal(t, "actor_dead", res1.Err.Error())

	time.Sleep(210 * time.Millisecond)
	worker.HandleCall(addBalanceMsgWithErr{5})
	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)
	worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, "balance_error", res.Err.Error())
	assert.Equal(t, 2, sup.state.(supState).children[worker.pid.id].restarts)
	res = worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, "balance_error", res.Err.Error())
	assert.Equal(t, 3, sup.state.(supState).children[worker.pid.id].restarts)
	res = worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	res = worker.HandleCall(addBalanceMsgWithErr{5})
	assert.Equal(t, "actor_dead", res1.Err.Error())
	//assert.Equal(t, "balance_error", res.Err.Error())
	//assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)

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
			Args: 56,
		},
		Messages: []MessageInterface{
			addBalanceMsg{},
			addBalanceMsgWithErr{},
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
type addBalanceMsgWithErr struct {
	inc int
}

func (m addBalanceMsgWithErr) Handle(state StateInterface) MessageReply {
	st := state.(testChildState)
	st.bal += m.inc
	if st.bal > 1 {

		return MessageReply{
			ActorReply: Reply{
				Response: st.bal,
				Err:      fmt.Errorf("balance_error"),
			},
			Stop:  true,
			Err:   fmt.Errorf("Balance error"),
			State: st,
		}
	}

	return MessageReply{
		State: st,
	}

}

func (m addBalanceMsgWithErr) GetType() string {
	return "add_balance_error"
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
