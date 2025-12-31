package mining

import (
	"context"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/logging"
)

// TaskFunc is a function that can be supervised.
type TaskFunc func(ctx context.Context)

// SupervisedTask represents a background task with restart capability.
type SupervisedTask struct {
	name          string
	task          TaskFunc
	restartDelay  time.Duration
	maxRestarts   int
	restartCount  int
	running       bool
	lastStartTime time.Time
	cancel        context.CancelFunc
	mu            sync.Mutex
}

// TaskSupervisor manages background tasks with automatic restart on failure.
type TaskSupervisor struct {
	tasks   map[string]*SupervisedTask
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	started bool
}

// NewTaskSupervisor creates a new task supervisor.
func NewTaskSupervisor() *TaskSupervisor {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskSupervisor{
		tasks:  make(map[string]*SupervisedTask),
		ctx:    ctx,
		cancel: cancel,
	}
}

// RegisterTask registers a task for supervision.
// The task will be automatically restarted if it exits or panics.
func (s *TaskSupervisor) RegisterTask(name string, task TaskFunc, restartDelay time.Duration, maxRestarts int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = &SupervisedTask{
		name:         name,
		task:         task,
		restartDelay: restartDelay,
		maxRestarts:  maxRestarts,
	}
}

// Start starts all registered tasks.
func (s *TaskSupervisor) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.mu.Unlock()

	s.mu.RLock()
	for name, task := range s.tasks {
		s.startTask(name, task)
	}
	s.mu.RUnlock()
}

// startTask starts a single supervised task.
func (s *TaskSupervisor) startTask(name string, st *SupervisedTask) {
	st.mu.Lock()
	if st.running {
		st.mu.Unlock()
		return
	}
	st.running = true
	st.lastStartTime = time.Now()

	taskCtx, taskCancel := context.WithCancel(s.ctx)
	st.cancel = taskCancel
	st.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
			}

			// Run the task with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						logging.Error("supervised task panicked", logging.Fields{
							"task":  name,
							"panic": r,
						})
					}
				}()
				st.task(taskCtx)
			}()

			// Check if we should restart
			st.mu.Lock()
			st.restartCount++
			shouldRestart := st.restartCount <= st.maxRestarts || st.maxRestarts < 0
			restartDelay := st.restartDelay
			st.mu.Unlock()

			if !shouldRestart {
				logging.Warn("supervised task reached max restarts", logging.Fields{
					"task":       name,
					"maxRestart": st.maxRestarts,
				})
				return
			}

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(restartDelay):
				logging.Info("restarting supervised task", logging.Fields{
					"task":         name,
					"restartCount": st.restartCount,
				})
			}
		}
	}()

	logging.Info("started supervised task", logging.Fields{"task": name})
}

// Stop stops all supervised tasks.
func (s *TaskSupervisor) Stop() {
	s.cancel()
	s.wg.Wait()

	s.mu.Lock()
	s.started = false
	for _, task := range s.tasks {
		task.mu.Lock()
		task.running = false
		task.mu.Unlock()
	}
	s.mu.Unlock()

	logging.Info("task supervisor stopped")
}

// GetTaskStatus returns the status of a task.
func (s *TaskSupervisor) GetTaskStatus(name string) (running bool, restartCount int, found bool) {
	s.mu.RLock()
	task, ok := s.tasks[name]
	s.mu.RUnlock()

	if !ok {
		return false, 0, false
	}

	task.mu.Lock()
	defer task.mu.Unlock()
	return task.running, task.restartCount, true
}

// GetAllTaskStatuses returns status of all tasks.
func (s *TaskSupervisor) GetAllTaskStatuses() map[string]TaskStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make(map[string]TaskStatus, len(s.tasks))
	for name, task := range s.tasks {
		task.mu.Lock()
		statuses[name] = TaskStatus{
			Name:         name,
			Running:      task.running,
			RestartCount: task.restartCount,
			LastStart:    task.lastStartTime,
		}
		task.mu.Unlock()
	}
	return statuses
}

// TaskStatus contains the status of a supervised task.
type TaskStatus struct {
	Name         string    `json:"name"`
	Running      bool      `json:"running"`
	RestartCount int       `json:"restartCount"`
	LastStart    time.Time `json:"lastStart"`
}
