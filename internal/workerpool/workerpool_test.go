package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockJob - mock-модель Job интерфейса для тестов
type MockJob struct {
	doFunc   func(ctx context.Context)
	doneChan chan struct{}
}

func (j *MockJob) Do(ctx context.Context) {
	if j.doFunc != nil {
		j.doFunc(ctx)
	}
	if j.doneChan != nil {
		close(j.doneChan)
	}
}

func TestWorkerPool_AddJob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		wp := NewWorkerPool(1)
		wp.Run()

		jobDone := make(chan struct{})
		job := &MockJob{doneChan: jobDone}

		err := wp.AddJob(job)
		assert.NoError(t, err)

		select {
		case <-jobDone:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	})

	t.Run("failed", func(t *testing.T) {
		wp := NewWorkerPool(1)
		wp.Run()
		require.NoError(t, wp.Close())

		err := wp.AddJob(&MockJob{})
		assert.Error(t, err)
		assert.Equal(t, "worker pool is closed", err.Error())
	})
}

func TestWorkerPool_Run(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		wp := NewWorkerPool(2)
		counter := atomic.Int32{}

		var wg sync.WaitGroup
		wg.Add(2)

		// Создаем 2 задачи, которые увеличивают счетчик
		job := &MockJob{
			doFunc: func(ctx context.Context) {
				counter.Add(1)
				wg.Done()
			},
		}

		// Добавляем задачи до запуска
		require.NoError(t, wp.AddJob(job))
		require.NoError(t, wp.AddJob(job))

		// Запускаем пул
		wp.Run()
		wp.Run() // второй вызов должен быть проигнорирован

		wg.Wait()
		assert.Equal(t, int32(2), counter.Load())
	})

	t.Run("cancel", func(t *testing.T) {
		wp := NewWorkerPool(1)
		jobDone := make(chan struct{})

		job := &MockJob{
			doFunc: func(ctx context.Context) {
				<-ctx.Done() // ждем отмены контекста
				close(jobDone)
			},
		}

		wp.Run()
		require.NoError(t, wp.AddJob(job))

		// Закрываем пул
		require.NoError(t, wp.Close())

		select {
		case <-jobDone:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("context is cancel")
		}
	})
}
