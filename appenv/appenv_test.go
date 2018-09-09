package appenv

import (
	"testing"
	"time"
)

// TestGetEnvChecksMutex is just a concurrency experiment
func TestGetEnvChecksMutex(t *testing.T) {
	waitduration := 3 * time.Second
	tstart := time.Now()
	go func() {
		envMutex.Lock()
		t.Log("Locked env mutex, now sleeping")
		time.Sleep(waitduration)
		envMutex.Unlock()
		t.Log("Unlocked env mutex")
	}()
	time.Sleep(500 * time.Millisecond) // wait for the goroutine to lock
	e := GetEnv()
	telapsed := time.Since(tstart)

	if telapsed < waitduration {
		t.Errorf("Read permitted of environment global var while RW lock open. Elapsed: %d, Env: %v", telapsed/time.Second, &e)
	}
}
