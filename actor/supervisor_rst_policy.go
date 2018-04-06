package actor

import "time"

type restartPolicy interface {
	setupMonitor(sup *Sup, child actorInterface)
}

type policyOneForOne struct {
}

type policyAllForOne struct{}

//type policyRestForOne struct {}

func (rp policyOneForOne) setupMonitor(sup *Sup, child actorInterface) {
	mon := newMonitor(sup, child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*Actor).spec.RestartRetryIn)
		return sup.(*Sup).restartChild(actr.(*Actor))
	})
}

func (rp policyAllForOne) setupMonitor(sup *Sup, child actorInterface) {
	mon := newMonitor(sup, child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*Actor).spec.RestartRetryIn)
		res := sup.(*Sup).restartChild(actr.(*Actor))

		return res
	})

}
