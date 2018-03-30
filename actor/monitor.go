package actor

type monitor struct {
	parent  actorInterface
	child   actorInterface
	errChan chan error
	isDead  chan error
}

func newMonitor(parent, child actorInterface) *monitor {
	return &monitor{
		parent:  parent,
		child:   child,
		errChan: make(chan error),
		isDead:  make(chan error, 1),
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
				m.setDead(restartErr)
			} else {
				m.setAlive()
			}
		}
	}()
}

func (m *monitor) trigger(err error) {
	m.errChan <- err
}

func (m *monitor) setDead(err error) {
	m.isDead <- err
}

func (m *monitor) setAlive() {
	m.isDead <- nil
}

func (m *monitor) waitRestart() error {
	return <-m.isDead
}
