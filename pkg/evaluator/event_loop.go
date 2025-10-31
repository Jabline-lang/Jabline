package evaluator

import (
	"sync"
	"time"

	"jabline/pkg/object"
)

type Task struct {
	Fn       func()
	Delay    time.Duration
	Promise  *object.Promise
	Callback func(object.Object)
}

type EventLoop struct {
	tasks      chan *Task
	timers     []*Timer
	running    bool
	wg         sync.WaitGroup
	mutex      sync.Mutex
	tickerDone chan bool
}

type Timer struct {
	ID       int
	Delay    time.Duration
	Callback func()
	Promise  *object.Promise
	Timer    *time.Timer
	Done     chan bool
}

func NewEventLoop() *EventLoop {
	return &EventLoop{
		tasks:      make(chan *Task, 1000),
		timers:     make([]*Timer, 0),
		running:    false,
		tickerDone: make(chan bool),
	}
}

func (el *EventLoop) Start() {
	el.mutex.Lock()
	if el.running {
		el.mutex.Unlock()
		return
	}
	el.running = true
	el.mutex.Unlock()

	el.wg.Add(1)
	go el.run()
}

func (el *EventLoop) Stop() {
	el.mutex.Lock()
	defer el.mutex.Unlock()

	if !el.running {
		return
	}

	el.running = false
	el.tickerDone <- true
	close(el.tasks)

	for _, timer := range el.timers {
		if timer.Timer != nil {
			timer.Timer.Stop()
		}
		if timer.Done != nil {
			close(timer.Done)
		}
	}

	el.wg.Wait()
}

func (el *EventLoop) run() {
	defer el.wg.Done()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-el.tasks:
			if !ok {
				return
			}
			el.executeTask(task)

		case <-ticker.C:
			continue

		case <-el.tickerDone:
			return
		}
	}
}

func (el *EventLoop) executeTask(task *Task) {
	if task.Delay > 0 {
		timer := time.AfterFunc(task.Delay, func() {
			task.Fn()
			if task.Callback != nil && task.Promise != nil {
				task.Callback(task.Promise.Value)
			}
		})

		timerObj := &Timer{
			Delay:   task.Delay,
			Promise: task.Promise,
			Timer:   timer,
		}

		el.mutex.Lock()
		el.timers = append(el.timers, timerObj)
		el.mutex.Unlock()
	} else {
		task.Fn()
		if task.Callback != nil && task.Promise != nil {
			task.Callback(task.Promise.Value)
		}
	}
}

func (el *EventLoop) ScheduleTask(fn func(), delay time.Duration) {
	if !el.running {
		el.Start()
	}

	task := &Task{
		Fn:    fn,
		Delay: delay,
	}

	select {
	case el.tasks <- task:
	default:
		go func() {
			if delay > 0 {
				time.Sleep(delay)
			}
			fn()
		}()
	}
}

func (el *EventLoop) SchedulePromiseTask(promise *object.Promise, fn func(), delay time.Duration) {
	if !el.running {
		el.Start()
	}

	task := &Task{
		Fn:      fn,
		Delay:   delay,
		Promise: promise,
	}

	select {
	case el.tasks <- task:
	default:
		go func() {
			if delay > 0 {
				time.Sleep(delay)
			}
			fn()
		}()
	}
}

func (el *EventLoop) SetTimeout(callback func(), delay time.Duration) *object.Promise {
	promise := object.NewPromise()

	if !el.running {
		el.Start()
	}

	task := &Task{
		Fn: func() {
			callback()
			promise.Resolve(NULL)
		},
		Delay:   delay,
		Promise: promise,
	}

	select {
	case el.tasks <- task:
	default:
		go func() {
			time.Sleep(delay)
			callback()
			promise.Resolve(NULL)
		}()
	}

	return promise
}

func (el *EventLoop) CreateResolvedPromise(value object.Object) *object.Promise {
	return object.NewResolvedPromise(value)
}

func (el *EventLoop) CreateRejectedPromise(reason object.Object) *object.Promise {
	return object.NewRejectedPromise(reason)
}

func (el *EventLoop) CreatePendingPromise() *object.Promise {
	return object.NewPromise()
}

func (el *EventLoop) Await(promise *object.Promise) (object.Object, object.Object) {
	switch promise.State {
	case object.RESOLVED:
		return promise.Value, nil
	case object.REJECTED:
		return nil, promise.Reason
	case object.PENDING:
		return NULL, nil
	}
	return NULL, nil
}

func (el *EventLoop) IsRunning() bool {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	return el.running
}

func (el *EventLoop) Wait() {
	el.wg.Wait()
}

var GlobalEventLoop = NewEventLoop()
