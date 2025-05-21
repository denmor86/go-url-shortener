package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	DefaultJobChanSize = 16
)

type Job interface {
	Do(ctx context.Context)
}

type WorkerPool struct {
	countWorkers int
	jobsChan     chan Job
	doneChan     chan struct{}
	closed       atomic.Bool
	processOnce  sync.Once
	wg           sync.WaitGroup
}

func NewWorkerPool(count int) *WorkerPool {
	return &WorkerPool{
		countWorkers: count,
		jobsChan:     make(chan Job, DefaultJobChanSize),
		doneChan:     make(chan struct{}),
	}
}

func (wp *WorkerPool) AddJob(job Job) error {
	if wp.closed.Load() {
		return fmt.Errorf("worker pool is closed")
	}
	// закидываем задачу в канал
	wp.jobsChan <- job

	return nil
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-wp.jobsChan:
			if !ok {
				return
			}

			job.Do(ctx)
		}
	}
}

func (wp *WorkerPool) Run() {
	if wp.closed.Load() {
		return
	}

	wp.processOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			// ожидание отмены
			<-wp.doneChan
			cancel()
		}()

		for range wp.countWorkers {
			// запускаем workerы
			wp.wg.Add(1)
			go wp.worker(ctx)
		}
	})
}

func (wp *WorkerPool) Close() error {
	isCanceled := wp.closed.Swap(true)

	if !isCanceled {
		close(wp.doneChan)
		close(wp.jobsChan)
	}

	return nil
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
