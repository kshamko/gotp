package actor

import "time"

type restartPolicy interface {
	setupMonitor(sup *Sup, child actorInterface)
}

type policyOneForOne struct {
}

type policyAllForOne struct {
}

type policyRestForOne struct {
}

func (rp policyOneForOne) setupMonitor(sup *Sup, child actorInterface) {
	mon := newMonitor(sup, child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*actor).spec.RestartRetryIn)
		return sup.(*Sup).supervisorRestartChild(actr.(*actor))
	})
}

func (rp policyAllForOne) setupMonitor(sup *Sup, child actorInterface) {
	mon := newMonitor(sup, child)
	mon.start(func(sup, actr actorInterface) error {
		time.Sleep(actr.(*actor).spec.RestartRetryIn)
		res := sup.(*Sup).supervisorRestartChild(actr.(*actor))

		return res
	})

}
