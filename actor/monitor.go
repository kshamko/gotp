package actor

type monitor struct {
	parent  actorInterface
	child   actorInterface
	errChan chan error
	dead    bool
}

func newMonitor(parent, child actorInterface) *monitor {
	return &monitor{
		parent:  parent,
		child:   child,
		errChan: make(chan error),
	}
}

func (m *monitor) start(restartFunc func(parent actorInterface, child actorInterface)) {
	m.child.setMonitor(m)
	go func() {
		err := <-m.errChan
		if err != nil {
			restartFunc(m.parent, m.child)
		}
		close(m.errChan)
	}()
}

func (m *monitor) trigger(err error) {
	//f !m.dead {
	go func(m *monitor) {
		m.errChan <- err
	}(m)
	//	m.dead = true
	//}
}
