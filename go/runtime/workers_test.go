package runtime

import (
	"testing"
	"time"

	appkitconfig "github.com/emersonary/appkit/config"
)

type testAppConfig struct {
	appkitconfig.BaseConfig
}

func (c testAppConfig) Infra() *appkitconfig.BaseConfig {
	return &c.BaseConfig
}

func TestRegisterWorkerCancelledOnStop(t *testing.T) {
	var app Application[testAppConfig]
	workerCtx, _ := app.RegisterWorker()

	done := make(chan struct{})
	go func() {
		<-workerCtx.Done()
		close(done)
	}()

	app.stopWorkers()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker context was not cancelled")
	}
}

func TestRegisterWorkerMultipleWorkers(t *testing.T) {
	var app Application[testAppConfig]
	ctx1, _ := app.RegisterWorker()
	ctx2, _ := app.RegisterWorker()

	app.stopWorkers()

	if ctx1.Err() == nil || ctx2.Err() == nil {
		t.Fatal("expected all registered workers to be cancelled")
	}
}

func TestRegisterWorkerEarlyCancel(t *testing.T) {
	var app Application[testAppConfig]
	workerCtx, cancel := app.RegisterWorker()
	cancel()

	if workerCtx.Err() == nil {
		t.Fatal("expected worker context to be cancelled")
	}

	app.stopWorkers()
}
