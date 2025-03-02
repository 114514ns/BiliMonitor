package main

import (
	"sync"
)

type Task func()

type Worker struct {
	taskQueue chan Task
	wg        sync.WaitGroup
}

func NewWorker() *Worker {
	w := &Worker{
		taskQueue: make(chan Task, 100),
	}
	w.wg.Add(1)
	go w.run()
	return w
}

func (w *Worker) run() {
	defer w.wg.Done()
	for task := range w.taskQueue {
		task() // 执行任务
	}
}

func (w *Worker) AddTask(task Task) {
	w.taskQueue <- task
}

func (w *Worker) Stop() {
	close(w.taskQueue)
	w.wg.Wait()
}
