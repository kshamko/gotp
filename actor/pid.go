package actor

import (
	"math/rand"
	"time"
)

//Pid exists for started actor
type pid struct {
	id        string
	startTime time.Time
	//countRestarts int
	//supervisor    *Pid
	//spec          ChildSpec
	//messages map[string]struct{}
}

func newPid() pid {
	return pid{
		startTime: time.Now(),
		id:        randStringBytesMaskImprSrc(16),
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randStringBytesMaskImprSrc(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

/*
//Call handles sync call to started actor
func (pid *Pid) Call(msg MessageInterface) Reply {

	//fmt.Println("PID:", pid)

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
	fmt.Println("WORKER_ACTOR:", pid)
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

func (pid *Pid) GetRestartCount() int {
	return pid.countRestarts
}

//
func (pid *Pid) getRestartIntensity() float64 {
	diff := time.Since(pid.startTime)
	return diff.Seconds() / float64(pid.countRestarts)
}*/
