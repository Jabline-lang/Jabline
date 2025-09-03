package evaluator

import (
	"sync"
	"time"

	"jabline/pkg/object"
)

// Task representa una tarea a ejecutar en el event loop
type Task struct {
	Fn       func()
	Delay    time.Duration
	Promise  *object.Promise
	Callback func(object.Object)
}

// EventLoop maneja la ejecución asíncrona
type EventLoop struct {
	tasks      chan *Task
	timers     []*Timer
	running    bool
	wg         sync.WaitGroup
	mutex      sync.Mutex
	tickerDone chan bool
}

// Timer representa un temporizador para setTimeout
type Timer struct {
	ID       int
	Delay    time.Duration
	Callback func()
	Promise  *object.Promise
	Timer    *time.Timer
	Done     chan bool
}

// NewEventLoop crea un nuevo event loop
func NewEventLoop() *EventLoop {
	return &EventLoop{
		tasks:      make(chan *Task, 1000),
		timers:     make([]*Timer, 0),
		running:    false,
		tickerDone: make(chan bool),
	}
}

// Start inicia el event loop
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

// Stop detiene el event loop
func (el *EventLoop) Stop() {
	el.mutex.Lock()
	defer el.mutex.Unlock()

	if !el.running {
		return
	}

	el.running = false
	el.tickerDone <- true
	close(el.tasks)

	// Cancelar todos los timers
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

// run es el bucle principal del event loop
func (el *EventLoop) run() {
	defer el.wg.Done()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-el.tasks:
			if !ok {
				return // Canal cerrado
			}
			el.executeTask(task)

		case <-ticker.C:
			// Procesamiento periódico si es necesario
			continue

		case <-el.tickerDone:
			return
		}
	}
}

// executeTask ejecuta una tarea
func (el *EventLoop) executeTask(task *Task) {
	if task.Delay > 0 {
		// Ejecutar después de un delay
		timer := time.AfterFunc(task.Delay, func() {
			task.Fn()
			if task.Callback != nil && task.Promise != nil {
				task.Callback(task.Promise.Value)
			}
		})

		// Guardar referencia al timer si es necesario
		timerObj := &Timer{
			Delay:   task.Delay,
			Promise: task.Promise,
			Timer:   timer,
		}

		el.mutex.Lock()
		el.timers = append(el.timers, timerObj)
		el.mutex.Unlock()
	} else {
		// Ejecutar inmediatamente
		task.Fn()
		if task.Callback != nil && task.Promise != nil {
			task.Callback(task.Promise.Value)
		}
	}
}

// ScheduleTask programa una tarea para ejecutar
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
		// Tarea programada exitosamente
	default:
		// Canal lleno, ejecutar sincrónicamente como fallback
		go func() {
			if delay > 0 {
				time.Sleep(delay)
			}
			fn()
		}()
	}
}

// SchedulePromiseTask programa una tarea que resuelve una Promise
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
		// Tarea programada exitosamente
	default:
		// Canal lleno, ejecutar sincrónicamente como fallback
		go func() {
			if delay > 0 {
				time.Sleep(delay)
			}
			fn()
		}()
	}
}

// SetTimeout implementa setTimeout JavaScript-style
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
		// Tarea programada exitosamente
	default:
		// Canal lleno, ejecutar sincrónicamente como fallback
		go func() {
			time.Sleep(delay)
			callback()
			promise.Resolve(NULL)
		}()
	}

	return promise
}

// CreateResolvedPromise crea una Promise ya resuelta
func (el *EventLoop) CreateResolvedPromise(value object.Object) *object.Promise {
	return object.NewResolvedPromise(value)
}

// CreateRejectedPromise crea una Promise ya rechazada
func (el *EventLoop) CreateRejectedPromise(reason object.Object) *object.Promise {
	return object.NewRejectedPromise(reason)
}

// CreatePendingPromise crea una Promise en estado pending
func (el *EventLoop) CreatePendingPromise() *object.Promise {
	return object.NewPromise()
}

// Await simula el comportamiento de await (para usar en el evaluador)
func (el *EventLoop) Await(promise *object.Promise) (object.Object, object.Object) {
	// En un intérprete real, esto sería más complejo
	// Por simplicidad, esperamos de forma bloqueante
	switch promise.State {
	case object.RESOLVED:
		return promise.Value, nil
	case object.REJECTED:
		return nil, promise.Reason
	case object.PENDING:
		// Para un MVP, retornamos inmediatamente
		// En una implementación completa, esto pausaría la ejecución
		return NULL, nil
	}
	return NULL, nil
}

// IsRunning verifica si el event loop está ejecutándose
func (el *EventLoop) IsRunning() bool {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	return el.running
}

// Wait espera a que todas las tareas pendientes se completen
func (el *EventLoop) Wait() {
	el.wg.Wait()
}

// Global event loop instance
var GlobalEventLoop = NewEventLoop()
