package actor

import (
	"fmt"
	"time"
)

func getWorketSpec() ChildSpec {
	return ChildSpec{
		IsSupervisor:   false,
		RestartCount:   2,
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
