package actor

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorUnderLoad(t *testing.T) {

	sup, _ := SupervisorStart(SupOneForOne)
	worker, _ := sup.StartChild(getWorkerSpec())

	w := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		w.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				worker.HandleCast(addBalanceMsg{1})
			}
			w.Done()
		}()
	}

	w.Wait()

	res := worker.HandleCall(getBalanceMsg{})

	assert.Equal(t, 10056, res.Response.(int))
}
