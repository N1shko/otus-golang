package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	var errCount int64
	var wg sync.WaitGroup
	var once sync.Once

	taskPool := make(chan Task)
	done := make(chan struct{})

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case task, ok := <-taskPool:
					if !ok {
						return
					}
					if err := task(); err != nil {
						if atomic.AddInt64(&errCount, 1) >= int64(m) {
							once.Do(func() {
								close(done)
							})
						}
					}
				}
			}
		}()
	}

	for _, task := range tasks {
		select {
		case <-done:
			close(taskPool)
			wg.Wait()
			return ErrErrorsLimitExceeded
		case taskPool <- task:
		}
	}

	close(taskPool)
	wg.Wait()

	if atomic.LoadInt64(&errCount) >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
