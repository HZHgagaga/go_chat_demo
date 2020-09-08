package hnet

import (
	"runtime"

	"github.com/jeanphorn/log4go"
)

var (
	WorkPool  *WorkThread
	AsyncPool *AsyncThreadPool
)

func init() {
	WorkPool = NewWorkThread()
	AsyncPool = NewAsyncThreadPool(runtime.NumCPU())
}

type Task func()

//单协程工作队列
type WorkThread struct {
	taskQueue chan Task
}

func NewWorkThread() *WorkThread {
	return &WorkThread{
		taskQueue: make(chan Task, 1000),
	}
}

func (t *WorkThread) Start() {
	go func() {
		for {
			select {
			case task := <-t.taskQueue:
				log4go.Debug("Get task")
				task()
			}
		}
	}()
	log4go.Debug("WorkThread start...")
}

func (t *WorkThread) AddTask(task Task) {
	log4go.Debug("之前", len(t.taskQueue))
	t.taskQueue <- task

	log4go.Debug("之后", len(t.taskQueue))
}

//协程池，用于异步处理IO操作
type AsyncThreadPool struct {
	taskQueue []chan Task
	threadNum int
	index     int
}

func NewAsyncThreadPool(num int) *AsyncThreadPool {
	return &AsyncThreadPool{
		taskQueue: make([]chan Task, num),
		threadNum: num,
		index:     0,
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
	log4go.Debug("AsyncThreadPool start...")
}

func (pool *AsyncThreadPool) AsyncRun(task Task) {
	if pool.index == 1000000 {
		pool.index = 0
	}

	pool.taskQueue[pool.index%pool.threadNum] <- task
}
