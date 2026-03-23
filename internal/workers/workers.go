package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	UUID      uuid.UUID
	ObjectKey string
}

type WorkerPoolConfig struct {
	WorkersCount int
	Timeout      time.Duration
	In           chan Task
	Fn           func(ctx context.Context, in Task, deps *dependencies.AppDeps)
}

type WorkerPoolOutConfig struct {
	WorkersCount int
	Timeout      time.Duration
	Tick         time.Duration
	Out          chan Task
	GlobalCtx    context.Context
	Fn           func(ctx context.Context, deps *dependencies.AppDeps) Task
}

func UpConsumerWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config WorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for range config.WorkersCount {
			go func() {
				defer wg.Done()
				ctx := context.Background()
				for task := range config.In {
					ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
					config.Fn(ctxTimeout, task, deps)
					cancel()
				}
			}()
		}
	}()
}

func UpProducerWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config WorkerPoolOutConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for range config.WorkersCount {
			go func() {
				defer wg.Done()

				ctx := context.Background()

				ticker := time.NewTicker(config.Tick)
				sem := make(chan struct{}, 1)
				for {
					sem <- struct{}{}

					select {
					case <-config.GlobalCtx.Done():
						return
					case <-ticker.C:
						go func() {
							defer func() { _ = <-sem }()
							ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
							defer cancel()
							task := config.Fn(ctxTimeout, deps)

							select {
							case <-config.GlobalCtx.Done():
								return
							case config.Out <- task:
							}
						}()
					}
				}
			}()
		}
	}()
}
