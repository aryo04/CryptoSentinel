package network

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// GoroutineFunc represents a function that runs in a goroutine
type GoroutineFunc func(ctx context.Context) error

// SupervisedGoroutine represents a goroutine managed by the supervisor
type SupervisedGoroutine struct {
	ID            string
	Name          string
	Function      GoroutineFunc
	RestartPolicy RestartPolicy
	
	// Runtime state
	running       int32 // atomic
	restartCount  int
	lastError     error
	lastRestart   time.Time
	ctx           context.Context
	cancel        context.CancelFunc
}

// RestartPolicy defines how a goroutine should be restarted
type RestartPolicy struct {
	MaxRestarts     int
	RestartDelay    time.Duration
	BackoffFactor   float64
	MaxBackoffDelay time.Duration
	OnFailure       func(error, int) // Called on failure with error and restart count
}

// DefaultRestartPolicy returns a default restart policy
func DefaultRestartPolicy() RestartPolicy {
	return RestartPolicy{
		MaxRestarts:     5,
		RestartDelay:    1 * time.Second,
		BackoffFactor:   2.0,
		MaxBackoffDelay: 30 * time.Second,
	}
}

// GoroutineSupervisor manages and supervises goroutines
type GoroutineSupervisor struct {
	goroutines map[string]*SupervisedGoroutine
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    int32 // atomic
}

// NewGoroutineSupervisor creates a new goroutine supervisor
func NewGoroutineSupervisor(ctx context.Context) *GoroutineSupervisor {
	if ctx == nil {
		ctx = context.Background()
	}
	
	supervisorCtx, cancel := context.WithCancel(ctx)
	
	return &GoroutineSupervisor{
		goroutines: make(map[string]*SupervisedGoroutine),
		ctx:        supervisorCtx,
		cancel:     cancel,
	}
}

// Register registers a new goroutine with the supervisor
func (gs *GoroutineSupervisor) Register(id, name string, fn GoroutineFunc, policy RestartPolicy) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	if _, exists := gs.goroutines[id]; exists {
		return fmt.Errorf("goroutine with ID %s already registered", id)
	}
	
	sg := &SupervisedGoroutine{
		ID:            id,
		Name:          name,
		Function:      fn,
		RestartPolicy: policy,
	}
	
	gs.goroutines[id] = sg
	
	log.Printf("👁️ Registered goroutine: %s (%s)", name, id)
	return nil
}

// Start starts the supervisor and all registered goroutines
func (gs *GoroutineSupervisor) Start() error {
	if !atomic.CompareAndSwapInt32(&gs.running, 0, 1) {
		return fmt.Errorf("supervisor already running")
	}
	
	gs.mu.RLock()
	goroutines := make([]*SupervisedGoroutine, 0, len(gs.goroutines))
	for _, sg := range gs.goroutines {
		goroutines = append(goroutines, sg)
	}
	gs.mu.RUnlock()
	
	// Start all goroutines
	for _, sg := range goroutines {
		gs.startGoroutine(sg)
	}
	
	log.Printf("👁️ Supervisor started with %d goroutines", len(goroutines))
	return nil
}

// Stop stops the supervisor and all goroutines
func (gs *GoroutineSupervisor) Stop() {
	if !atomic.CompareAndSwapInt32(&gs.running, 1, 0) {
		return
	}
	
	log.Println("👁️ Stopping supervisor...")
	
	// Cancel context to signal all goroutines to stop
	gs.cancel()
	
	// Cancel individual goroutine contexts
	gs.mu.RLock()
	for _, sg := range gs.goroutines {
		if sg.cancel != nil {
			sg.cancel()
		}
	}
	gs.mu.RUnlock()
	
	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		gs.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Println("👁️ All goroutines stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("⚠️ Timeout waiting for goroutines to stop")
	}
	
	log.Println("👁️ Supervisor stopped")
}

// startGoroutine starts a supervised goroutine
func (gs *GoroutineSupervisor) startGoroutine(sg *SupervisedGoroutine) {
	if atomic.LoadInt32(&sg.running) == 1 {
		return
	}
	
	// Create context for this goroutine
	sg.ctx, sg.cancel = context.WithCancel(gs.ctx)
	atomic.StoreInt32(&sg.running, 1)
	
	gs.wg.Add(1)
	go gs.runGoroutine(sg)
	
	log.Printf("▶️ Started goroutine: %s", sg.Name)
}

// runGoroutine runs a goroutine with supervision
func (gs *GoroutineSupervisor) runGoroutine(sg *SupervisedGoroutine) {
	defer gs.wg.Done()
	defer atomic.StoreInt32(&sg.running, 0)
	
	for {
		// Check if supervisor is stopping
		if atomic.LoadInt32(&gs.running) == 0 {
			return
		}
		
		// Run the goroutine function
		err := sg.Function(sg.ctx)
		
		// Check if context was cancelled (normal shutdown)
		select {
		case <-sg.ctx.Done():
			log.Printf("⏹️ Goroutine %s stopped (context cancelled)", sg.Name)
			return
		default:
		}
		
		// Handle error
		if err != nil {
			sg.lastError = err
			sg.restartCount++
			
			log.Printf("❌ Goroutine %s failed (restart %d/%d): %v",
				sg.Name, sg.restartCount, sg.RestartPolicy.MaxRestarts, err)
			
			// Call failure handler if provided
			if sg.RestartPolicy.OnFailure != nil {
				sg.RestartPolicy.OnFailure(err, sg.restartCount)
			}
			
			// Check if we should restart
			if sg.restartCount > sg.RestartPolicy.MaxRestarts {
				log.Printf("💀 Goroutine %s exceeded max restarts, giving up", sg.Name)
				return
			}
			
			// Calculate backoff delay
			delay := gs.calculateBackoff(sg)
			sg.lastRestart = time.Now().Add(delay)
			
			log.Printf("🔄 Restarting goroutine %s in %v", sg.Name, delay)
			
			// Wait before restarting
			select {
			case <-time.After(delay):
				// Continue loop to restart
			case <-sg.ctx.Done():
				return
			}
		} else {
			// Goroutine exited normally without error
			log.Printf("✅ Goroutine %s completed successfully", sg.Name)
			return
		}
	}
}

// calculateBackoff calculates the backoff delay for a restart
func (gs *GoroutineSupervisor) calculateBackoff(sg *SupervisedGoroutine) time.Duration {
	delay := sg.RestartPolicy.RestartDelay
	
	// Apply exponential backoff
	for i := 1; i < sg.restartCount; i++ {
		delay = time.Duration(float64(delay) * sg.RestartPolicy.BackoffFactor)
		if delay > sg.RestartPolicy.MaxBackoffDelay {
			delay = sg.RestartPolicy.MaxBackoffDelay
			break
		}
	}
	
	return delay
}

// RestartGoroutine manually restarts a specific goroutine
func (gs *GoroutineSupervisor) RestartGoroutine(id string) error {
	gs.mu.RLock()
	sg, exists := gs.goroutines[id]
	gs.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("goroutine with ID %s not found", id)
	}
	
	// Stop the goroutine
	if sg.cancel != nil {
		sg.cancel()
	}
	
	// Wait a moment for it to stop
	time.Sleep(100 * time.Millisecond)
	
	// Reset restart count for manual restart
	sg.restartCount = 0
	
	// Start it again
	gs.startGoroutine(sg)
	
	return nil
}

// GetStatus returns the status of all supervised goroutines
func (gs *GoroutineSupervisor) GetStatus() map[string]GoroutineStatus {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	status := make(map[string]GoroutineStatus)
	
	for id, sg := range gs.goroutines {
		status[id] = GoroutineStatus{
			ID:           sg.ID,
			Name:         sg.Name,
			Running:      atomic.LoadInt32(&sg.running) == 1,
			RestartCount: sg.restartCount,
			LastError:    sg.lastError,
			LastRestart:  sg.lastRestart,
		}
	}
	
	return status
}

// GoroutineStatus represents the status of a supervised goroutine
type GoroutineStatus struct {
	ID           string
	Name         string
	Running      bool
	RestartCount int
	LastError    error
	LastRestart  time.Time
}

// IsHealthy checks if all goroutines are healthy
func (gs *GoroutineSupervisor) IsHealthy() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	for _, sg := range gs.goroutines {
		if atomic.LoadInt32(&sg.running) == 0 {
			return false
		}
		if sg.restartCount > sg.RestartPolicy.MaxRestarts/2 {
			return false // Consider unhealthy if restarted too many times
		}
	}
	
	return true
}

// GetMetrics returns supervisor metrics
func (gs *GoroutineSupervisor) GetMetrics() SupervisorMetrics {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	metrics := SupervisorMetrics{
		TotalGoroutines: len(gs.goroutines),
	}
	
	for _, sg := range gs.goroutines {
		if atomic.LoadInt32(&sg.running) == 1 {
			metrics.RunningGoroutines++
		} else {
			metrics.StoppedGoroutines++
		}
		metrics.TotalRestarts += sg.restartCount
	}
	
	return metrics
}

// SupervisorMetrics contains supervisor metrics
type SupervisorMetrics struct {
	TotalGoroutines   int
	RunningGoroutines int
	StoppedGoroutines int
	TotalRestarts     int
}