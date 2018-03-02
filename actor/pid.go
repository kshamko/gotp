package actor

import (
	"fmt"
)

//Pid exists for started actor
type Pid struct {
	a  *actor
	ID string
}

//Call handles sync call to started actor
func (pid *Pid) Call(msg MessageInterface) Reply {
	if _, ok := pid.a.messages[msg.GetType()]; ok {
		return pid.a.handleCall(msg)
	}

	return Reply{
		Err:      fmt.Errorf("unknown_msg_type"),
		Response: nil,
	}
}

//Cast handles async call to started actor
func (pid *Pid) Cast(msg MessageInterface) Reply {
	if _, ok := pid.a.messages[msg.GetType()]; ok {
		return pid.a.handleCast(msg)
	}

	return Reply{
		Err:      fmt.Errorf("unknown_msg_type"),
		Response: nil,
	}
}

//GetStopReason for actor
func (pid *Pid) GetStopReason() error {
	return pid.a.getStopReason()
}
