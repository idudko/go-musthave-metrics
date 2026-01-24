package agent

import (
	"context"
	"log"
	"sync"
)

type Task func(ctx context.Context) error

type WorkerPool struct {
	workerCount int
	tasks       chan Task
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		tasks:       make(chan Task, 100),
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *WorkerPool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}

func (p *WorkerPool) EnqueueTask(task Task) {
	p.tasks <- task
}

func (p *WorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}

			if err := task(ctx); err != nil {
				log.Printf("Error executing task: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
