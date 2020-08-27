package hnet

import (
	"fmt"
	"math/rand"
	"time"
)

type Task func()

//单协程工作队列
type WorkThread struct {
	TaskQueue chan Task
}

func NewWorkThread() *WorkThread {
	return &WorkThread{
		TaskQueue: make(chan Task, 5000),
	}
}

func (t *WorkThread) Start() {
	go func() {
		for {
			select {
			case task := <-t.TaskQueue:
				task()
			}
		}
	}()
	fmt.Println("WorkThread start...")
}

func (t *WorkThread) AddTask(task Task) {
	t.TaskQueue <- task
}

//协程池，用于异步处理IO操作
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

func (pool *AsyncThreadPool) asyncThreadRunFunc(num int) {
	pool.taskQueue[num] = make(chan Task, 5000)
	for {
		select {
		case task := <-pool.taskQueue[num]:
			task()
		}
	}
}

func (pool *AsyncThreadPool) Start() {
	for i := 0; i < pool.threadNum; i++ {
		go pool.asyncThreadRunFunc(i)
	}
	fmt.Println("AsyncThreadPool start...")
}

func (pool *AsyncThreadPool) AsyncRun(task Task) {
	//取随机数来负载均衡协程调度
	r := pool.rand.Intn(pool.threadNum)
	pool.taskQueue[r] <- task
}
