package actor

type monitor struct {
	parent  actorInterface
	child   actorInterface
	errChan chan error
}

func newMonitor(parent, child actorInterface) *monitor {
	return &monitor{
		parent:  parent,
		child:   child,
		errChan: make(chan error),
	}
}

func (m *monitor) start(restartFunc func(parent actorInterface, child actorInterface) error) {
	m.child.setMonitor(m)
	go func() {
		err := <-m.errChan
		close(m.errChan)

		if err != nil {
			restartErr := restartFunc(m.parent, m.child)
			if restartErr != nil {
				m.child.setDead(restartErr)
			} else {
				m.child.setAlive()
			}
		}
	}()
}

func (m *monitor) trigger(err error) {
	go func(m *monitor) {
		m.errChan <- err
	}(m)
}
