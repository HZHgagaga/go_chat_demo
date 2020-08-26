package hnet

import (
	"fmt"
	"math/rand"
	"time"
)

type Task func()

type WorkThread struct {
	TaskQueue chan Task
}

func NewWorkThread() *WorkThread {
	return &WorkThread{
		TaskQueue: make(chan Task, 5000),
	}
}

func (t *WorkThread) Start() {
	fmt.Println("WorkThread start...")
	go func() {
		for {
			select {
			case task := <-t.TaskQueue:
				task()
			}
		}
	}()
}

func (t *WorkThread) AddTask(task Task) {
	t.TaskQueue <- task
}

type AsyncThreadPool struct {
	taskQueue []chan Task
	threadNum int
	rand      *rand.Rand
}

func NewAsyncThreadPool(num int) *AsyncThreadPool {
	return &AsyncThreadPool{
		taskQueue: make([]chan Task, num),
		threadNum: num,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (pool *AsyncThreadPool) Start() {
	for i := 0; i < pool.threadNum; i++ {
		go func(num int) {
			for {
				select {
				case task := <-pool.taskQueue[num]:
					task()
				}
			}
		}(i)
	}
}

func (pool *AsyncThreadPool) AsyncRun(task Task) {
	r := pool.rand.Intn(pool.threadNum)
	pool.taskQueue[r] <- task
}
