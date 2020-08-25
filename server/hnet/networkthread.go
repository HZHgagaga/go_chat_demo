package hnet

import "fmt"

type Task func()

type WorkThread struct {
	TaskQueue chan Task
}

func NewWorkThread() *WorkThread {
	return &WorkThread{
		TaskQueue: make(chan Task, 5000),
	}
}

func (pool *WorkThread) Start() {
	fmt.Println("WorkThread start...")
	go func() {
		for {
			select {
			case task := <-pool.TaskQueue:
				fmt.Println("Run task")
				task()
				fmt.Println("Task end")
			}
		}
	}()
}

func (pool *WorkThread) AddTask(task Task) {
	pool.TaskQueue <- task
}
