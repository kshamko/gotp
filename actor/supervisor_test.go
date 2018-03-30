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
	worker.WaitRestart()
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)

	worker.HandleCall(addBalanceMsgWithErr{2})
	res = worker.HandleCall(addBalanceMsgWithErr{2})
	assert.Equal(t, ErrRestarting, res.Err)

	worker.WaitRestart()
	assert.Equal(t, 2, sup.state.(supState).children[worker.pid.id].restarts)

	res = worker.HandleCall(addBalanceMsgWithErr{2})
	assert.Equal(t, "balance_error", res.Err.Error())

	err := worker.WaitRestart()
	assert.Equal(t, ErrDead, err)
}

func TestSupervisorRecoverChild(t *testing.T) {
	sup, _ := SupervisorStart(SupOneForAll)
	worker, _ := sup.SupervisorStartChild(getWorketSpec())

	worker.HandleCall(addBalanceMsgWithErr{2})
	worker.WaitRestart()

	worker.HandleCall(addBalanceMsgWithErr{2})
	worker.WaitRestart()
	assert.Equal(t, 2, sup.state.(supState).children[worker.pid.id].restarts)

	time.Sleep(30 * time.Millisecond)
	worker.HandleCall(addBalanceMsgWithErr{2})
	worker.WaitRestart()
	assert.Equal(t, 1, sup.state.(supState).children[worker.pid.id].restarts)
}
