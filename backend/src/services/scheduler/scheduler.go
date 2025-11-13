package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task represents a scheduled task
type Task struct {
	Name     string
	Interval time.Duration
	Fn       func(context.Context) error
}

// Scheduler manages background tasks
type Scheduler struct {
	tasks   []*Task
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:   make([]*Task, 0),
		running: false,
	}
}

// AddTask adds a task to the scheduler
func (s *Scheduler) AddTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = append(s.tasks, task)
	log.Printf("Scheduler: Added task '%s' with interval %v", task.Name, task.Interval)
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true

	log.Printf("Scheduler: Starting with %d tasks", len(s.tasks))

	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(task)
	}

	log.Println("Scheduler: Started successfully")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}

	log.Println("Scheduler: Stopping...")
	s.cancel()
	s.running = false
	s.mu.Unlock()

	// Wait for all tasks to complete
	s.wg.Wait()

	log.Println("Scheduler: Stopped successfully")
	return nil
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// runTask runs a single task on its schedule
func (s *Scheduler) runTask(task *Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	log.Printf("Scheduler: Task '%s' started (interval: %v)", task.Name, task.Interval)

	// Run immediately on start
	if err := s.executeTask(task); err != nil {
		log.Printf("Scheduler: Task '%s' failed on initial run: %v", task.Name, err)
	}

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Scheduler: Task '%s' stopped", task.Name)
			return
		case <-ticker.C:
			if err := s.executeTask(task); err != nil {
				log.Printf("Scheduler: Task '%s' failed: %v", task.Name, err)
			}
		}
	}
}

// executeTask executes a task with timeout protection
func (s *Scheduler) executeTask(task *Task) error {
	// Create a timeout context for the task
	taskCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	log.Printf("Scheduler: Executing task '%s'", task.Name)
	startTime := time.Now()

	err := task.Fn(taskCtx)

	duration := time.Since(startTime)
	if err != nil {
		log.Printf("Scheduler: Task '%s' completed with error in %v: %v", task.Name, duration, err)
	} else {
		log.Printf("Scheduler: Task '%s' completed successfully in %v", task.Name, duration)
	}

	return err
}

// GetTaskCount returns the number of registered tasks
func (s *Scheduler) GetTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

// Global scheduler instance
var (
	globalScheduler *Scheduler
	schedulerMu     sync.Mutex
)

// Initialize creates and starts the global scheduler
func Initialize() error {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if globalScheduler != nil {
		return nil
	}

	globalScheduler = NewScheduler()
	log.Println("Scheduler: Initialized")
	return nil
}

// GetScheduler returns the global scheduler instance
func GetScheduler() *Scheduler {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	return globalScheduler
}

// StartGlobal starts the global scheduler
func StartGlobal() error {
	scheduler := GetScheduler()
	if scheduler == nil {
		return nil
	}
	return scheduler.Start()
}

// StopGlobal stops the global scheduler
func StopGlobal() error {
	scheduler := GetScheduler()
	if scheduler == nil {
		return nil
	}
	return scheduler.Stop()
}
