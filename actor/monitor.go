package actor

type monitor struct {
	parent  *actor
	child   *actor
	errChan chan error
}

func newMonitor(parent, child *actor) *monitor {
	return &monitor{
		parent:  parent,
		child:   child,
		errChan: make(chan error),
	}
}

func (m *monitor) start(restartFunc func(m *monitor)) {

	go func() {
		err := <-m.errChan
		close(m.errChan)
		if err != nil {
			//m.sup.supervisorRestartChild(m.child)
			restartFunc(m)
		}
	}()

}
