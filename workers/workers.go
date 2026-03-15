package workers

import (
	"context"
	"time"
)

type Worker interface {
	Config() UpConfig
	Work()
}

type UpConfig struct {
	JobTick      time.Duration
	IdleSleep    time.Duration
	WorkersCount int
	JobCount     int
	JobTimeout   time.Duration
	SignalCTX    context.Context
	FailCount    int
}

func Up(worker Worker) {
	t := time.NewTicker(worker.Config().JobTick)

	for {
		select {
		case <-t.C:
			for i := 0; i < worker.Config().WorkersCount; i++ {
				go worker.Work()
			}
		case <-worker.Config().SignalCTX.Done():
			return
		}
	}
}
