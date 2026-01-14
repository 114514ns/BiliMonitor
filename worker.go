package main

import (
	"sync"
)

type WorkerTask func()

type Worker struct {
	taskQueue chan WorkerTask
	wg        sync.WaitGroup
}

func NewWorker(poolSize int) *Worker {
	w := &Worker{
		taskQueue: make(chan WorkerTask, 10000),
	}
	for i := 0; i < poolSize; i++ {
		w.wg.Add(1)
		go w.run()
	}
	return w
}
func (w *Worker) QueueLen() int {
	return len(w.taskQueue)
}
func (w *Worker) run() {
	defer w.wg.Done()
	for task := range w.taskQueue {
		task()
	}
}

func (w *Worker) AddTask(task WorkerTask) {
	w.taskQueue <- task
}

func (w *Worker) Stop() {
	close(w.taskQueue)
	w.wg.Wait()
}
