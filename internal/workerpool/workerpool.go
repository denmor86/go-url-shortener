package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Внутренние константы воркера
const (
	// defaultJobChanSize значение по-умолчанию размера канала  для воркера
	defaultJobChanSize = 16
)

// Job - интерфейс задачи
type Job interface {
	Do(ctx context.Context)
}

// WorkerPool - структура вокер пула
type WorkerPool struct {
	countWorkers int           // количество воркеров
	jobsChan     chan Job      // канал с задачами
	doneChan     chan struct{} // канала управления
	closed       atomic.Bool   // признак остановки пула
	processOnce  sync.Once
	wg           sync.WaitGroup
}

// NewWorkerPool - метод создания воркер пула
func NewWorkerPool(count int) *WorkerPool {
	return &WorkerPool{
		countWorkers: count,
		jobsChan:     make(chan Job, defaultJobChanSize),
		doneChan:     make(chan struct{}),
	}
}

// AddJob - метод добавления задачи в канал задач
func (wp *WorkerPool) AddJob(job Job) error {
	if wp.closed.Load() {
		return fmt.Errorf("worker pool is closed")
	}
	// закидываем задачу в канал
	wp.jobsChan <- job

	return nil
}

// worker- метод управления воркером
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

// Run - метод запуска воркера
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

// Close - метод закрытия каналов воркера
func (wp *WorkerPool) Close() error {
	isCanceled := wp.closed.Swap(true)

	if !isCanceled {
		close(wp.doneChan)
		close(wp.jobsChan)
	}

	return nil
}

// Wait - ожидание
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
