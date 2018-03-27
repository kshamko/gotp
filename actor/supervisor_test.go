//go test -cover -coverprofile=c.out ./actor
//go tool cover -html=c.out

package actor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupervisorStart(t *testing.T) {
	sup, err := SupervisorStart(SupOneForAll)
	assert.Nil(t, err)
	assert.NotEmpty(t, sup.pid.id)
	assert.Equal(t, 0, len(sup.state.(supState).children))
}

func TestSupervisorStartChild(t *testing.T) {
	sup, _ := SupervisorStart(SupOneForAll)

	worker, err := sup.SupervisorStartChild(getWorketSpec())

	assert.Nil(t, err)
	assert.NotEmpty(t, worker.pid.id)
	assert.Equal(t, 1, len(sup.state.(supState).children))
}

func TestSupervisorRestartChild(t *testing.T) {
	sup, _ := SupervisorStart(SupOneForAll)
	worker, _ := sup.SupervisorStartChild(getWorketSpec())
	assert.Equal(t, 0, sup.state.(supState).children[worker.pid.id].restarts)

	res := worker.HandleCall(addBalanceMsgWithErr{2})
	worker.HandleCall(addBalanceMsgWithErr{2})
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)

	res1 := worker.HandleCall(addBalanceMsgWithErr{5})
	assert.Equal(t, "actor_dead", res1.Err.Error())

	time.Sleep(210 * time.Millisecond)
	worker.HandleCall(addBalanceMsgWithErr{5})
	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)
	worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, "balance_error", res.Err.Error())
	assert.Equal(t, 2, sup.state.(supState).children[worker.pid.id].restarts)
	res = worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	assert.Equal(t, "balance_error", res.Err.Error())
	assert.Equal(t, 3, sup.state.(supState).children[worker.pid.id].restarts)
	res = worker.HandleCall(addBalanceMsgWithErr{5})

	time.Sleep(105 * time.Millisecond)
	res = worker.HandleCall(addBalanceMsgWithErr{5})
	assert.Equal(t, "actor_dead", res1.Err.Error())
	//assert.Equal(t, "balance_error", res.Err.Error())
	//assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)

}
