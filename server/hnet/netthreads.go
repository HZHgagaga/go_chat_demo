package hnet

import "fmt"

type Task func()

type ThreadPool struct {
	TaskQueue chan Task
}

func NewThreadPool() *ThreadPool {
	return &ThreadPool{
		TaskQueue: make(chan Task, 5000),
	}
}

func (pool *ThreadPool) Start(num int) {
	fmt.Println("ThreadPool start...")

	for i := 0; i < num; i++ {
		go func() {
			select {
			case task := <-pool.TaskQueue:
				task()
			}
		}()
	}
}

func (pool *ThreadPool) AddTask(task Task) {
	pool.TaskQueue <- task
}
